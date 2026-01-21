package places

import (
	"context"
	"redcat/internal/model"
)

type storage interface {
	Upsert(ctx context.Context, p model.Place) error
}

type Service struct {
	store storage
}

func New(s storage) *Service {
	return &Service{
		store: s,
	}
}

func (s *Service) Add(ctx context.Context, p model.Place) error {
	return s.store.Upsert(ctx, p)
}
