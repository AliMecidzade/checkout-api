package service

import (
	"context"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"
	"checkout-api/internal/validation"
)

type CartService struct {
	cart    repository.CartRepo
	catalog repository.CatalogRepo
}

func NewCartService(cart repository.CartRepo, catalog repository.CatalogRepo) *CartService {
	return &CartService{
		cart:    cart,
		catalog: catalog,
	}
}

func (s *CartService) UpsertItem(ctx context.Context, userID int64, itemID int64, quantity int64) error {
	var verr validation.ValidationError
	if quantity <= 0 {
		verr.With("quantity", "gt=0", "quantity must be greater than zero")
	}
	if len(verr.Fields) != 0 {
		return &verr
	}

	if _, err := s.catalog.GetItem(ctx, itemID); err != nil {
		return fmt.Errorf("get item %d: %w", itemID, err)
	}

	return s.cart.UpsertCartItem(ctx, userID, itemID, quantity)
}

func (s *CartService) List(ctx context.Context, userID int64, p domain.Page) ([]domain.CartItem, bool, error) {
	p = normalizePage(p)

	items, hasMore, err := s.cart.ListCart(ctx, userID, p)
	if err != nil {
		return nil, false, fmt.Errorf("list cart: %w", err)
	}

	return items, hasMore, nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, itemID int64) error {
	return s.cart.RemoveCartItem(ctx, userID, itemID)
}
