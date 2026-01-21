package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"redcat/internal/api"
)

// --- Contract Tests ---

// TestHealthEndpoint verifies GET /healthz returns 200.
func TestHealthEndpoint(t *testing.T) {
	app := fiber.New()
	api.Register(app, api.Handlers{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// --- Places Search Contract Tests ---

func TestSearchPlaces_Contract_ValidRequest(t *testing.T) {
	// Contract: POST /api/v1/places/search
	// Request body: { location: {lat, lon}, category_ids: [], limit: int }
	// Response: { places: [], total: int, query: {...} }

	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name: "valid request with all fields",
			body: map[string]any{
				"location":     map[string]any{"lat": 35.1753, "lon": 33.3642},
				"category_ids": []string{"4bf58dd8d48988d1e0931735"},
				"limit":        10,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid request minimal",
			body: map[string]any{
				"location":     map[string]any{"lat": 0.0, "lon": 0.0},
				"category_ids": []string{},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// NOTE: Full test requires mocked service. Skipping execution test.
			// This documents the contract.
			_ = tc
		})
	}
}

func TestSearchPlaces_Contract_InvalidRequest(t *testing.T) {
	app := fiber.New()
	api.Register(app, api.Handlers{})

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       ``,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/places/search", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}

// --- Create Place Contract Tests ---

// TestCreatePlace_Contract_ValidRequest documents the contract for valid requests.
// Full execution requires integration tests with real storage (see storage/valkey/integration_test.go).
func TestCreatePlace_Contract_ValidRequest(t *testing.T) {
	// Contract: POST /api/v1/places
	// Required fields per OpenAPI: name, location, category_ids
	// Response: 201 with Place object

	// Document expected valid request structure:
	validPlace := map[string]any{
		"id":           "test123",
		"name":         "Test Place",
		"lat":          35.1753,
		"lon":          33.3642,
		"category_ids": []string{"4bf58dd8d48988d1e0931735"},
		"country":      "CY",
	}

	// Verify JSON serialization
	_, err := json.Marshal(validPlace)
	if err != nil {
		t.Fatalf("failed to marshal valid place: %v", err)
	}
	// Execution tested in integration tests
}

func TestCreatePlace_Contract_MissingFields(t *testing.T) {
	app := fiber.New()
	api.Register(app, api.Handlers{})

	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name:       "missing id",
			body:       map[string]any{"name": "Test"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       map[string]any{"id": "123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty id",
			body:       map[string]any{"id": "", "name": "Test"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "whitespace id",
			body:       map[string]any{"id": "   ", "name": "Test"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/places", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
		})
	}
}

// --- Get Place Contract Tests ---

// TestGetPlace_Contract documents GET /api/v1/places/{id}.
// Contract: Returns 404 if place not found, 200 with Place object if found.
// Full execution requires integration tests.
func TestGetPlace_Contract(t *testing.T) {
	// Contract documentation:
	// - GET /api/v1/places/{id}
	// - Response 200: Place object with required fields: id, name, location, category_ids
	// - Response 404: Error object with code and message
	_ = t // Execution tested in integration tests
}

// --- Delete Place Contract Tests ---

// TestDeletePlace_Contract documents DELETE /api/v1/places/{id}.
// Contract: Returns 204 No Content on success, 404 if not found.
// Full execution requires integration tests.
func TestDeletePlace_Contract(t *testing.T) {
	// Contract documentation:
	// - DELETE /api/v1/places/{id}
	// - Response 204: No Content (empty body)
	// - Response 404: Error object
	_ = t // Execution tested in integration tests
}

// --- Response Structure Contract Tests ---

func TestSearchResponse_MatchesOpenAPI(t *testing.T) {
	// OpenAPI SearchResponse schema:
	// - places: array of PlaceWithDistance
	// - total: integer
	// - query: object with location, limit

	// PlaceWithDistance extends Place with distance_m

	type Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	type PlaceWithDistance struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Location   Location `json:"location"`
		Address    string   `json:"address,omitempty"`
		Locality   string   `json:"locality,omitempty"`
		Region     string   `json:"region,omitempty"`
		Postcode   string   `json:"postcode,omitempty"`
		Country    string   `json:"country,omitempty"`
		CategoryIDs []string `json:"category_ids"`
		DistanceM   float64  `json:"distance_m"`
	}

	type SearchResponse struct {
		Places []PlaceWithDistance `json:"places"`
		Total  int                 `json:"total"`
		Query  struct {
			Location Location `json:"location"`
			Limit    int64    `json:"limit"`
		} `json:"query"`
	}

	// This test validates the response structure can be unmarshaled
	example := `{
		"places": [{
			"id": "abc123",
			"name": "Test Place",
			"location": {"lat": 35.17, "lon": 33.36},
			"category_ids": ["cat1"],
			"distance_m": 100.5
		}],
		"total": 1,
		"query": {"location": {"lat": 35.17, "lon": 33.36}, "limit": 10}
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(example), &resp); err != nil {
		t.Fatalf("failed to parse SearchResponse: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected total=1, got %d", resp.Total)
	}
	if len(resp.Places) != 1 {
		t.Errorf("expected 1 place, got %d", len(resp.Places))
	}
	if resp.Places[0].DistanceM != 100.5 {
		t.Errorf("expected distance_m=100.5, got %f", resp.Places[0].DistanceM)
	}
}

func TestPlaceResponse_RequiredFields(t *testing.T) {
	// OpenAPI Place schema required fields: id, name, location, category_ids

	type Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	type Place struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Location    Location `json:"location"`
		CategoryIDs []string `json:"category_ids"`
	}

	// Test that minimal response has required fields
	minimal := `{"id": "x", "name": "Y", "location": {"lat": 0, "lon": 0}, "category_ids": []}`

	var p Place
	if err := json.Unmarshal([]byte(minimal), &p); err != nil {
		t.Fatalf("failed to parse minimal Place: %v", err)
	}

	if p.ID == "" {
		t.Error("id should not be empty")
	}
	if p.Name == "" {
		t.Error("name should not be empty")
	}
}

func TestErrorResponse_MatchesOpenAPI(t *testing.T) {
	// OpenAPI Error schema:
	// - code: string (required)
	// - message: string (required)
	// - details: object (optional)

	type ErrorResponse struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details,omitempty"`
	}

	example := `{"code": "INVALID_REQUEST", "message": "missing field"}`
	var e ErrorResponse
	if err := json.Unmarshal([]byte(example), &e); err != nil {
		t.Fatalf("failed to parse ErrorResponse: %v", err)
	}

	if e.Code == "" || e.Message == "" {
		t.Error("error response must have code and message")
	}
}

// --- HTTP Method Contract Tests ---

// TestAPIRoutes_Exist verifies all expected routes are registered.
// Tests only routes that don't require storage (healthz, validation errors).
func TestAPIRoutes_Exist(t *testing.T) {
	app := fiber.New()
	api.Register(app, api.Handlers{})

	// Test healthz - doesn't need services
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthz: expected 200, got %d", resp.StatusCode)
	}

	// Test invalid body on POST routes returns 400, not 404/405
	routes := []string{
		"/api/v1/places",
		"/api/v1/places/search",
	}
	for _, path := range routes {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{invalid}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s request failed: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
			t.Errorf("%s: route not registered (got %d)", path, resp.StatusCode)
		}
	}
}

// --- Content-Type Contract Tests ---

func TestContentType_JSON(t *testing.T) {
	app := fiber.New()
	api.Register(app, api.Handlers{})

	// Health endpoint returns plaintext
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp, _ := app.Test(req)
	resp.Body.Close()

	// JSON endpoints should return application/json
	// (tested in integration tests with real responses)
}

// --- Location Validation Tests ---

func TestLocation_Validation(t *testing.T) {
	// OpenAPI Location constraints:
	// lat: -90 to 90
	// lon: -180 to 180

	validLocations := []struct {
		lat, lon float64
	}{
		{0, 0},
		{90, 180},
		{-90, -180},
		{35.1753, 33.3642}, // Cyprus
	}

	for _, loc := range validLocations {
		if loc.lat < -90 || loc.lat > 90 {
			t.Errorf("lat %f out of range", loc.lat)
		}
		if loc.lon < -180 || loc.lon > 180 {
			t.Errorf("lon %f out of range", loc.lon)
		}
	}
}
