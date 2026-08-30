package postgres

import (
	"context"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
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

func (s *PostgresStore) ListCart(ctx context.Context, userID int64, p domain.Page) ([]domain.CartItem, bool, error) {
	var (
		rows pgx.Rows
		err  error
	)
	limit := int64(p.Limit) + 1
	if p.Cursor != nil {
		rows, err = s.DB().GetItemsFromUserCartAfterID(ctx, userID, limit, p.Cursor.ID)
	} else {
		rows, err = s.DB().GetItemsFromUserCartOffset(ctx, userID, limit, int64(p.Offset))
	}
	if err != nil {
		return nil, false, fmt.Errorf("get cart for user %d: %w", userID, err)
	}
	defer rows.Close()

	items := make([]domain.CartItem, 0, p.Limit)
	for rows.Next() {
		var it domain.CartItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.Price, &it.Stock, &it.CreatedAt, &it.Quantity); err != nil {
			return nil, false, fmt.Errorf("scan cart item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate cart: %w", err)
	}

	hasMore := int64(len(items)) > int64(p.Limit)
	if hasMore {
		items = items[:p.Limit]
	}
	return items, hasMore, nil
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
