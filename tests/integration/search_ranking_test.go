// Package integration contains integration tests for the RedCat API.
//
//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// Paphos center point (Paphos Archaeological Park area)
const (
	paphosLat = 34.7575
	paphosLon = 32.4070
)

// TestSearchRanking tests search results are correctly ranked by distance.
// Uses 100 real POIs from the Paphos, Cyprus area as fixture data.
func TestSearchRanking(t *testing.T) {
	// Load fixture
	places := loadPaphosFixture(t)
	if len(places) < 10 {
		t.Fatalf("need at least 10 places in fixture, got %d", len(places))
	}

	// Add test prefix to avoid conflicts
	testPrefix := "search-rank-test-"
	for i := range places {
		places[i].ID = testPrefix + places[i].ID
	}

	// Create all places
	t.Log("Creating", len(places), "places...")
	for _, p := range places {
		createPlace(t, p)
	}

	// Cleanup after test
	t.Cleanup(func() {
		t.Log("Cleaning up test places...")
		for _, p := range places {
			deletePlace(t, p.ID)
		}
	})

	// Calculate expected distances from Paphos center
	type placeWithDist struct {
		Place            Place
		ExpectedDistance float64 // meters
	}
	var placesWithDist []placeWithDist
	for _, p := range places {
		dist := haversineDistance(paphosLat, paphosLon, p.Lat, p.Lon)
		placesWithDist = append(placesWithDist, placeWithDist{
			Place:            p,
			ExpectedDistance: dist,
		})
	}

	// Sort by expected distance
	sort.Slice(placesWithDist, func(i, j int) bool {
		return placesWithDist[i].ExpectedDistance < placesWithDist[j].ExpectedDistance
	})

	t.Logf("Closest place: %s at %.0fm", placesWithDist[0].Place.Name, placesWithDist[0].ExpectedDistance)
	t.Logf("Farthest place: %s at %.0fm", placesWithDist[len(placesWithDist)-1].Place.Name, placesWithDist[len(placesWithDist)-1].ExpectedDistance)

	// Search for top 10
	searchReq := SearchRequest{
		Location: Location{Lat: paphosLat, Lon: paphosLon},
		Limit:    10,
	}
	body, _ := json.Marshal(searchReq)
	resp, err := httpClient.Post(apiURL("/places/search"), "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("search request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, respBody)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Search returned %d results", len(result.Places))

	// Filter results to only our test places
	var testResults []PlaceWithDistance
	for _, p := range result.Places {
		if len(p.ID) > len(testPrefix) && p.ID[:len(testPrefix)] == testPrefix {
			testResults = append(testResults, p)
		}
	}

	if len(testResults) == 0 {
		t.Fatal("no test places found in search results")
	}

	t.Logf("Found %d test places in results", len(testResults))

	// Verify results are sorted by distance (ascending)
	t.Run("ResultsSortedByDistance", func(t *testing.T) {
		for i := 1; i < len(testResults); i++ {
			if testResults[i].DistanceM < testResults[i-1].DistanceM {
				t.Errorf("results not sorted: place %d (%.0fm) comes after place %d (%.0fm)",
					i, testResults[i].DistanceM, i-1, testResults[i-1].DistanceM)
			}
		}
	})

	// Verify API distances match expected haversine distances (within tolerance)
	t.Run("DistancesAccurate", func(t *testing.T) {
		tolerance := 100.0 // 100 meters tolerance for ECEF approximation

		for _, apiPlace := range testResults {
			// Find corresponding fixture place
			originalID := apiPlace.ID[len(testPrefix):]
			var fixturePlace *Place
			for _, p := range places {
				if p.ID == testPrefix+originalID {
					fixturePlace = &p
					break
				}
			}
			if fixturePlace == nil {
				t.Errorf("could not find fixture place for %s", apiPlace.ID)
				continue
			}

			expectedDist := haversineDistance(paphosLat, paphosLon, fixturePlace.Lat, fixturePlace.Lon)
			diff := math.Abs(apiPlace.DistanceM - expectedDist)

			if diff > tolerance {
				t.Errorf("distance mismatch for %s: API=%.0fm, expected=%.0fm, diff=%.0fm",
					apiPlace.Name, apiPlace.DistanceM, expectedDist, diff)
			} else {
				t.Logf("✓ %s: API=%.0fm, expected=%.0fm, diff=%.0fm",
					apiPlace.Name, apiPlace.DistanceM, expectedDist, diff)
			}
		}
	})

	// Verify top results match expected top places
	t.Run("TopResultsCorrect", func(t *testing.T) {
		// Check that the closest places from our fixture appear in results
		expectedTopIDs := make(map[string]bool)
		for i := 0; i < min(10, len(placesWithDist)); i++ {
			expectedTopIDs[placesWithDist[i].Place.ID] = true
		}

		matchCount := 0
		for _, p := range testResults {
			if expectedTopIDs[p.ID] {
				matchCount++
			}
		}

		// All returned test results should be from expected top
		if matchCount != len(testResults) {
			t.Errorf("expected all %d results to be from top places, got %d matches",
				len(testResults), matchCount)
		}
	})
}

// haversineDistance calculates the distance in meters between two points.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func loadPaphosFixture(t *testing.T) []Place {
	t.Helper()

	// Find testdata relative to this file
	_, filename, _, _ := runtime.Caller(0)
	testdataPath := filepath.Join(filepath.Dir(filename), "testdata", "paphos_places.jsonl")

	f, err := os.Open(testdataPath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	var places []Place
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var p Place
		if err := json.Unmarshal(scanner.Bytes(), &p); err != nil {
			t.Fatalf("failed to parse fixture line: %v", err)
		}
		places = append(places, p)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	return places
}

func createPlace(t *testing.T, p Place) {
	t.Helper()

	body, _ := json.Marshal(p)
	resp, err := httpClient.Post(apiURL("/places"), "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create place failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("create place %s failed: %d: %s", p.ID, resp.StatusCode, respBody)
	}
}

func deletePlace(t *testing.T, id string) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodDelete, apiURL("/places/"+id), nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Logf("delete place %s failed: %v", id, err)
		return
	}
	defer resp.Body.Close()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
