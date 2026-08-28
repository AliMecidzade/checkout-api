package postgres

import (
	"context"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"
)

func (s *PostgresStore) UpsertCartItem(ctx context.Context, userID int64, itemID int64, quantity int64) error {
	if _, err := s.DB().UpsertCart(ctx, userID, itemID, quantity); err != nil {
		if pgCode(err) == "23503" {
			return fmt.Errorf("item %d: %w", itemID, repository.ErrNotFound)
		}
		return fmt.Errorf("upsert cart item %d for user %d: %w", itemID, userID, err)
	}
	return nil
}

func (s *PostgresStore) GetUserCart(ctx context.Context, userID int64) ([]domain.CartItem, error) {
	rows, err := s.DB().GetItemsFromUserCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get cart for user %d: %w", userID, err)
	}
	defer rows.Close()

	items := make([]domain.CartItem, 0)
	for rows.Next() {
		var it domain.CartItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.Price, &it.Stock, &it.CreatedAt, &it.Quantity); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) DeleteUserCart(ctx context.Context, userID int64) error {
	if _, err := s.DB().DeleteCartByUserID(ctx, userID); err != nil {
		return fmt.Errorf("delete cart for user %d: %w", userID, err)
	}
	return nil
}

func (s *PostgresStore) RemoveCartItem(ctx context.Context, userID int64, itemID int64) error {
	tag, err := s.DB().DeleteItemFromUserCart(ctx, userID, itemID)
	if err != nil {
		return fmt.Errorf("remove cart item %d for user %d: %w", itemID, userID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("item %d in cart of user %d: %w", itemID, userID, repository.ErrNotFound)
	}
	return nil
}
