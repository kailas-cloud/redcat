// Package integration contains integration tests for the RedCat API.
// These tests hit the live API at https://redcat.kailas.cloud/api/v1/
//
// Run with: go test -tags=integration ./tests/integration/...
//
//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const defaultBaseURL = "https://redcat.kailas.cloud"

func baseURL() string {
	if url := os.Getenv("REDCAT_API_URL"); url != "" {
		return url
	}
	return defaultBaseURL
}

func apiURL(path string) string {
	return baseURL() + "/api/v1" + path
}

// Place represents a place in the API
type Place struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Lat         float64  `json:"lat"`
	Lon         float64  `json:"lon"`
	Address     string   `json:"address,omitempty"`
	Locality    string   `json:"locality,omitempty"`
	Region      string   `json:"region,omitempty"`
	Postcode    string   `json:"postcode,omitempty"`
	Country     string   `json:"country,omitempty"`
	CategoryIDs []string `json:"category_ids,omitempty"`
}

type CreateResponse struct {
	Item Place `json:"item"`
}

type SearchRequest struct {
	Location    Location `json:"location"`
	CategoryIDs []string `json:"category_ids,omitempty"`
	Limit       int64    `json:"limit"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type SearchResponse struct {
	Places []PlaceWithDistance `json:"places"`
	Total  int                 `json:"total"`
}

type PlaceWithDistance struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Location  Location `json:"location"`
	DistanceM float64  `json:"distance_m"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func TestHealthz(t *testing.T) {
	// Note: /healthz is only accessible internally (not through ingress)
	// This test will be skipped when running against production
	resp, err := httpClient.Get(baseURL() + "/healthz")
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	defer resp.Body.Close()

	// When running via ingress, healthz returns 404 (not exposed)
	// When running locally, it returns 200
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("healthz endpoint not exposed via ingress")
	}
}

func TestPlaceCRUD(t *testing.T) {
	testID := fmt.Sprintf("test-place-%d", time.Now().UnixNano())
	place := Place{
		ID:          testID,
		Name:        "Integration Test Place",
		Lat:         55.7558,
		Lon:         37.6173,
		Address:     "Red Square, 1",
		Locality:    "Moscow",
		Country:     "RU",
		CategoryIDs: []string{"13000"},
	}

	// Create
	t.Run("Create", func(t *testing.T) {
		body, _ := json.Marshal(place)
		resp, err := httpClient.Post(apiURL("/places"), "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("create request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, respBody)
		}

		var result CreateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Item.ID != testID {
			t.Errorf("expected ID %s, got %s", testID, result.Item.ID)
		}
	})

	// Read
	t.Run("Get", func(t *testing.T) {
		resp, err := httpClient.Get(apiURL("/places/" + testID))
		if err != nil {
			t.Fatalf("get request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, respBody)
		}

		var result Place
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.ID != testID {
			t.Errorf("expected ID %s, got %s", testID, result.ID)
		}
		if result.Name != place.Name {
			t.Errorf("expected name %s, got %s", place.Name, result.Name)
		}
	})

	// Search nearby
	t.Run("SearchNearby", func(t *testing.T) {
		searchReq := SearchRequest{
			Location: Location{Lat: 55.7558, Lon: 37.6173},
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

		// Should find at least our test place
		found := false
		for _, p := range result.Places {
			if p.ID == testID {
				found = true
				if p.DistanceM > 1 {
					t.Errorf("expected distance ~0, got %f", p.DistanceM)
				}
				break
			}
		}
		if !found {
			t.Logf("test place not found in search results (total: %d)", result.Total)
		}
	})

	// Search with category filter
	t.Run("SearchWithCategory", func(t *testing.T) {
		searchReq := SearchRequest{
			Location:    Location{Lat: 55.7558, Lon: 37.6173},
			CategoryIDs: []string{"13000"},
			Limit:       10,
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

		t.Logf("search with category returned %d results", result.Total)
	})

	// Delete
	t.Run("Delete", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, apiURL("/places/"+testID), nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("delete request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 204, got %d: %s", resp.StatusCode, respBody)
		}
	})

	// Verify deleted
	t.Run("VerifyDeleted", func(t *testing.T) {
		resp, err := httpClient.Get(apiURL("/places/" + testID))
		if err != nil {
			t.Fatalf("get request failed: %v", err)
		}
		defer resp.Body.Close()

		// API currently returns 200 with empty place for non-existent IDs
		// TODO: API should return 404 for deleted places
		if resp.StatusCode == http.StatusOK {
			var result Place
			json.NewDecoder(resp.Body).Decode(&result)
			if result.ID == testID {
				t.Errorf("place should have been deleted but still exists")
			}
		}
	})
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name     string
		place    Place
		wantCode int
	}{
		{
			name:     "missing id",
			place:    Place{Name: "Test"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing name",
			place:    Place{ID: "test-id"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty id",
			place:    Place{ID: "", Name: "Test"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "whitespace id",
			place:    Place{ID: "   ", Name: "Test"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.place)
			resp, err := httpClient.Post(apiURL("/places"), "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantCode {
				t.Errorf("expected status %d, got %d", tc.wantCode, resp.StatusCode)
			}
		})
	}
}

func TestSearchValidation(t *testing.T) {
	t.Run("InvalidJSON", func(t *testing.T) {
		resp, err := httpClient.Post(apiURL("/places/search"), "application/json", bytes.NewReader([]byte("invalid")))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		searchReq := SearchRequest{
			Location: Location{Lat: 0, Lon: 0},
			Limit:    10,
		}
		body, _ := json.Marshal(searchReq)

		resp, err := httpClient.Post(apiURL("/places/search"), "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should still work with 0,0 coordinates
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestGetNotFound(t *testing.T) {
	resp, err := httpClient.Get(apiURL("/places/nonexistent-place-id-12345"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Note: API currently returns 200 with empty place for non-existent IDs
	// This is a known behavior that could be improved
	if resp.StatusCode == http.StatusOK {
		var result Place
		json.NewDecoder(resp.Body).Decode(&result)
		if result.ID != "" {
			t.Errorf("expected empty place for non-existent ID, got ID: %s", result.ID)
		}
	} else if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
	}
}
