package service

import (
	"context"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"
)

type CatalogService struct {
	items repository.CatalogRepo
}

func NewCatalogService(items repository.CatalogRepo) *CatalogService {
	return &CatalogService{items: items}
}

func (s *CatalogService) ListItems(ctx context.Context, p domain.Page) ([]domain.Item, bool, error) {
	p = normalizePage(p)

	items, hasMore, err := s.items.ListItems(ctx, p)
	if err != nil {
		return nil, false, fmt.Errorf("list items: %w", err)
	}
	return items, hasMore, nil
}

func (s *CatalogService) GetItem(ctx context.Context, id int64) (domain.Item, error) {
	return s.items.GetItem(ctx, id)
}
