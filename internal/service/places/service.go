package places

import (
	"context"
	"redcat/internal/domain/model"
	"redcat/internal/storage/valkey"
)

type Service struct {
	store *valkey.PlacesStorage
}

func New(store *valkey.PlacesStorage) *Service { return &Service{store: store} }

func (s *Service) Add(ctx context.Context, p model.Place) error {
	return s.store.Upsert(ctx, p)
}

func (s *Service) Get(ctx context.Context, id string) (model.Place, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

type SearchParams struct {
	Lat, Lon float64
	Limit    int64
	CategoryIDs []string
}

type SearchResult struct {
	Place     model.Place `json:"place"`
	DistanceM float64     `json:"distance_m"`
}

func (s *Service) SearchNearest(ctx context.Context, sp SearchParams) ([]SearchResult, error) {
	res, err := s.store.SearchNearest(ctx, valkey.SearchParams{
		Lat: sp.Lat, Lon: sp.Lon, Limit: sp.Limit, CategoryIDs: sp.CategoryIDs,
	})
	if err != nil { return nil, err }
	out := make([]SearchResult, 0, len(res))
	for _, r := range res {
		out = append(out, SearchResult{Place: r.Place, DistanceM: r.DistanceM})
	}
	return out, nil
}
