package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/parquet-go/parquet-go"
)

// ParquetPlace matches Foursquare parquet schema
type ParquetPlace struct {
	FsqPlaceID        string   `parquet:"fsq_place_id,optional"`
	Name              string   `parquet:"name,optional"`
	Latitude          float64  `parquet:"latitude,optional"`
	Longitude         float64  `parquet:"longitude,optional"`
	Address           string   `parquet:"address,optional"`
	Locality          string   `parquet:"locality,optional"`
	Region            string   `parquet:"region,optional"`
	Postcode          string   `parquet:"postcode,optional"`
	AdminRegion       string   `parquet:"admin_region,optional"`
	PostTown          string   `parquet:"post_town,optional"`
	PoBox             string   `parquet:"po_box,optional"`
	Country           string   `parquet:"country,optional"`
	DateCreated       string   `parquet:"date_created,optional"`
	DateRefreshed     string   `parquet:"date_refreshed,optional"`
	DateClosed        string   `parquet:"date_closed,optional"`
	Tel               string   `parquet:"tel,optional"`
	Website           string   `parquet:"website,optional"`
	Email             string   `parquet:"email,optional"`
	FacebookID        *int64   `parquet:"facebook_id,optional"`
	Instagram         string   `parquet:"instagram,optional"`
	Twitter           string   `parquet:"twitter,optional"`
	FsqCategoryIDs    []string `parquet:"fsq_category_ids,list,optional"`
	FsqCategoryLabels []string `parquet:"fsq_category_labels,list,optional"`
	PlacemakerURL     string   `parquet:"placemaker_url,optional"`
}

// APIPlace is the JSON body for POST /api/v1/places
type APIPlace struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Lat            float64  `json:"lat"`
	Lon            float64  `json:"lon"`
	Address        string   `json:"address,omitempty"`
	Locality       string   `json:"locality,omitempty"`
	Region         string   `json:"region,omitempty"`
	Postcode       string   `json:"postcode,omitempty"`
	AdminRegion    string   `json:"admin_region,omitempty"`
	PostTown       string   `json:"post_town,omitempty"`
	PoBox          string   `json:"po_box,omitempty"`
	Country        string   `json:"country,omitempty"`
	DateCreated    string   `json:"date_created,omitempty"`
	DateRefreshed  string   `json:"date_refreshed,omitempty"`
	DateClosed     string   `json:"date_closed,omitempty"`
	Tel            string   `json:"tel,omitempty"`
	Website        string   `json:"website,omitempty"`
	Email          string   `json:"email,omitempty"`
	FacebookID     string   `json:"facebook_id,omitempty"`
	Instagram      string   `json:"instagram,omitempty"`
	Twitter        string   `json:"twitter,omitempty"`
	CategoryIDs    []string `json:"category_ids"`
	CategoryLabels []string `json:"category_labels,omitempty"`
	PlacemakerURL  string   `json:"placemaker_url,omitempty"`
}

var (
	parquetFile = flag.String("file", "", "parquet file to load")
	apiURL      = flag.String("api", "https://redcat.kailas.cloud", "API base URL")
	workers     = flag.Int("workers", 10, "concurrent HTTP workers")
	limit       = flag.Int("limit", 0, "max records to load (0 = all)")
	dryRun      = flag.Bool("dry-run", false, "don't send to API")
	slim        = flag.Bool("slim", false, "only send id, name, lat, lon, category_ids, country")
)

func main() {
	flag.Parse()

	if *parquetFile == "" {
		log.Fatal("--file required")
	}

	// Open parquet file with generic reader
	f, err := os.Open(*parquetFile)
	if err != nil {
		log.Fatalf("open file: %v", err)
	}
	defer f.Close()

	reader := parquet.NewGenericReader[ParquetPlace](f)
	defer reader.Close()

	numRows := int(reader.NumRows())
	log.Printf("Parquet file: %s, rows: %d", *parquetFile, numRows)

	if *limit > 0 && *limit < numRows {
		numRows = *limit
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// HTTP client with connection pooling
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *workers * 2,
			MaxIdleConnsPerHost: *workers * 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Worker pool
	placeCh := make(chan ParquetPlace, *workers*10)
	var wg sync.WaitGroup
	var totalLoaded, totalErrors int64

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range placeCh {
				if err := sendPlace(ctx, client, p); err != nil {
					atomic.AddInt64(&totalErrors, 1)
				} else {
					atomic.AddInt64(&totalLoaded, 1)
				}
			}
		}()
	}

	// Progress reporter
	start := time.Now()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				loaded := atomic.LoadInt64(&totalLoaded)
				errors := atomic.LoadInt64(&totalErrors)
				elapsed := time.Since(start)
				rate := float64(loaded) / elapsed.Seconds()
				log.Printf("Progress: %d/%d (%.1f%%), rate: %.0f rec/s, errors: %d",
					loaded, numRows, float64(loaded)*100/float64(numRows), rate, errors)
			}
		}
	}()

	// Read and dispatch
	batchSize := 1000
	rowsRead := 0

	for {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		if *limit > 0 && rowsRead >= *limit {
			break
		}

		places := make([]ParquetPlace, batchSize)
		n, err := reader.Read(places)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			log.Printf("read error: %v", err)
			break
		}

		for i := 0; i < n; i++ {
			if *limit > 0 && rowsRead >= *limit {
				break
			}
			select {
			case <-ctx.Done():
				goto done
			case placeCh <- places[i]:
				rowsRead++
			}
		}

		if err == io.EOF {
			break
		}
	}

done:
	close(placeCh)
	wg.Wait()

	elapsed := time.Since(start)
	loaded := atomic.LoadInt64(&totalLoaded)
	errors := atomic.LoadInt64(&totalErrors)
	log.Printf("Done: %d records in %v (%.0f rec/s), errors: %d",
		loaded, elapsed, float64(loaded)/elapsed.Seconds(), errors)
}

func fmtFacebookID(id *int64) string {
	if id == nil || *id == 0 {
		return ""
	}
	return fmt.Sprintf("%d", *id)
}

func sendPlace(ctx context.Context, client *http.Client, p ParquetPlace) error {
	if p.FsqPlaceID == "" || p.Name == "" {
		return nil // skip invalid
	}

	if *dryRun {
		return nil
	}

	var ap APIPlace
	if *slim {
		// Minimal fields only: ~200-300 bytes/record
		ap = APIPlace{
			ID:          p.FsqPlaceID,
			Name:        p.Name,
			Lat:         p.Latitude,
			Lon:         p.Longitude,
			CategoryIDs: p.FsqCategoryIDs,
			Country:     p.Country,
		}
	} else {
		ap = APIPlace{
			ID:             p.FsqPlaceID,
			Name:           p.Name,
			Lat:            p.Latitude,
			Lon:            p.Longitude,
			Address:        p.Address,
			Locality:       p.Locality,
			Region:         p.Region,
			Postcode:       p.Postcode,
			AdminRegion:    p.AdminRegion,
			PostTown:       p.PostTown,
			PoBox:          p.PoBox,
			Country:        p.Country,
			DateCreated:    p.DateCreated,
			DateRefreshed:  p.DateRefreshed,
			DateClosed:     p.DateClosed,
			Tel:            p.Tel,
			Website:        p.Website,
			Email:          p.Email,
			FacebookID:     fmtFacebookID(p.FacebookID),
			Instagram:      p.Instagram,
			Twitter:        p.Twitter,
			CategoryIDs:    p.FsqCategoryIDs,
			CategoryLabels: p.FsqCategoryLabels,
			PlacemakerURL:  p.PlacemakerURL,
		}
	}

	body, _ := json.Marshal(ap)
	req, err := http.NewRequestWithContext(ctx, "POST", *apiURL+"/api/v1/places", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
