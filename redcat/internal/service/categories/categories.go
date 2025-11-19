package categories

import (
	"context"
	"errors"
	"redcat/internal/model"
	"strings"
)

type embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type categoryStorage interface {
	LooksAlike(ctx context.Context, vec []float32, k int64) ([]model.Category, error)
}

type Service struct {
	embedder embedder
	store    categoryStorage
}

func NewCategoryService(e embedder, s categoryStorage) *Service {
	return &Service{
		embedder: e,
		store:    s,
	}
}

func (s *Service) Search(ctx context.Context, query string, limit int64) ([]model.Category, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("empty query")
	}

	vec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	return s.store.LooksAlike(ctx, vec, limit)
}
