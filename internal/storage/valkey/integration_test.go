package valkey

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"redcat/internal/domain/model"
)

func getEnvAddrs() []string {
	v := os.Getenv("VALKEY_ADDRS")
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts { p = strings.TrimSpace(p); if p != "" { out = append(out, p) } }
	return out
}

func TestIntegration_Storage_KNN(t *testing.T) {
	addrs := getEnvAddrs()
	if len(addrs) == 0 {
		t.Skip("VALKEY_ADDRS not set; skipping integration test")
	}
	user := os.Getenv("VALKEY_USER")
	pass := os.Getenv("VALKEY_PASS")
	t.Logf("addrs=%v user=%q pass_len=%d", addrs, user, len(pass))

	cli, err := NewClient(addrs, user, pass)
	if err != nil { t.Fatalf("client: %v", err) }
	defer cli.Close()

	idx := "idx:itest:" + time.Now().Format("20060102T150405.000000000")
	prefix := "itest:" + time.Now().Format("150405.000000") + ":"

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := EnsurePlacesIndex(ctx, cli.R, idx, prefix); err != nil {
		t.Fatalf("EnsurePlacesIndex: %v", err)
	}

	s := NewPlacesStorage(cli.R, idx, prefix)
	// seed
	seed := []model.Place{
		{ID: "a", Name: "A", Lat: 35.1700, Lon: 33.3600, CategoryIDs: []string{"testcat"}},
		{ID: "b", Name: "B", Lat: 35.1710, Lon: 33.3610, CategoryIDs: []string{"testcat"}},
		{ID: "c", Name: "C", Lat: 51.5074, Lon: -0.1278, CategoryIDs: []string{"othercat"}}, // London
	}
	for _, p := range seed {
		if err := s.Upsert(ctx, p); err != nil {
			t.Fatalf("Upsert %s: %v", p.ID, err)
		}
	}
	// allow index to catch up
	time.Sleep(500 * time.Millisecond)

	res, err := s.SearchNearest(ctx, SearchParams{Lat: 35.1705, Lon: 33.3605, Limit: 2, CategoryIDs: []string{"testcat"}})
	if err != nil { t.Fatalf("SearchNearest: %v", err) }
	if len(res) == 0 { t.Fatalf("expected >0 results") }
	if res[0].Place.ID != "a" && res[0].Place.ID != "b" { t.Fatalf("unexpected first id: %s", res[0].Place.ID) }
	if len(res) == 2 && res[0].DistanceM > res[1].DistanceM {
		t.Fatalf("distance not sorted ascending: %f > %f", res[0].DistanceM, res[1].DistanceM)
	}

	// cleanup keys
	for _, p := range seed { _ = s.Delete(ctx, p.ID) }
}
