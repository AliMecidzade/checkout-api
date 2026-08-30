package postgres

import (
	"context"
	"errors"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) ListItems(ctx context.Context, p domain.Page) ([]domain.Item, bool, error) {
	var (
		rows pgx.Rows
		err  error
	)
	limit := int64(p.Limit) + 1
	if p.Cursor != nil {
		rows, err = s.DB().GetItemsAfterID(ctx, limit, p.Cursor.ID)
	} else {
		rows, err = s.DB().GetItemsOffset(ctx, limit, int64(p.Offset))
	}
	if err != nil {
		return nil, false, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0, p.Limit)
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.Price, &it.Stock, &it.CreatedAt); err != nil {
			return nil, false, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate items: %w", err)
	}

	hasMore := int64(len(items)) > int64(p.Limit)
	if hasMore {
		items = items[:p.Limit]
	}
	return items, hasMore, nil
}

func (s *PostgresStore) GetItem(ctx context.Context, id int64) (domain.Item, error) {
	var it domain.Item
	err := s.DB().GetItemByID(ctx, id).Scan(&it.ID, &it.Name, &it.Description, &it.Price, &it.Stock, &it.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item %d: %w", id, err)
	}
	return it, nil
}
