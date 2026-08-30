package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) CreateOrder(ctx context.Context, userID int64, items []domain.LineItem, total int64, status string) (domain.Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.WithTx(tx)

	for _, li := range items {
		var it domain.Item
		err := q.GetItemByIDForUpdate(ctx, li.ItemID).Scan(&it.ID, &it.Name, &it.Description, &it.Price, &it.Stock, &it.CreatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, fmt.Errorf("item %d: %w", li.ItemID, repository.ErrNotFound)
		}
		if err != nil {
			return domain.Order{}, fmt.Errorf("lock item %d: %w", li.ItemID, err)
		}
		if it.Stock < li.Quantity {
			return domain.Order{}, fmt.Errorf("item %d: %w", li.ItemID, repository.ErrInsufficient)
		}
	}

	var orderID int64
	var orderCreatedAt time.Time
	if err := q.InsertOrderReturning(ctx, userID, total, status).Scan(&orderID, &orderCreatedAt); err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	for _, li := range items {
		if _, err := q.InsertLineItem(ctx, orderID, li.ItemID, li.Price, li.Quantity); err != nil {
			return domain.Order{}, fmt.Errorf("insert line item for item %d: %w", li.ItemID, err)
		}
		if _, err := q.DecrementItemStock(ctx, li.ItemID, li.Quantity); err != nil {
			return domain.Order{}, fmt.Errorf("decrement stock for item %d: %w", li.ItemID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("commit transaction: %w", err)
	}

	return domain.Order{
		ID:        orderID,
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    status,
		CreatedAt: orderCreatedAt,
	}, nil
}

func (s *PostgresStore) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	if _, err := s.DB().UpdateOrderStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("update order %d status: %w", orderID, err)
	}
	return nil
}

func (s *PostgresStore) ListOrdersByUser(ctx context.Context, userID int64, p domain.Page) ([]domain.Order, bool, error) {
	var (
		rows pgx.Rows
		err  error
	)
	limit := int64(p.Limit) + 1
	if p.Cursor != nil {
		rows, err = s.DB().GetUserOrdersAfterID(ctx, userID, limit, p.Cursor.ID)
	} else {
		rows, err = s.DB().GetUserOrdersOffset(ctx, userID, limit, int64(p.Offset))
	}
	if err != nil {
		return nil, false, fmt.Errorf("get orders for user %d: %w", userID, err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0, p.Limit)
	orderIDs := make([]int64, 0, p.Limit)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, false, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
		orderIDs = append(orderIDs, o.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate orders: %w", err)
	}

	hasMore := int64(len(orders)) > int64(p.Limit)
	if hasMore {
		orders = orders[:p.Limit]
		orderIDs = orderIDs[:p.Limit]
	}

	if len(orderIDs) == 0 {
		return orders, hasMore, nil
	}

	itemsByOrder, err := s.lineItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, false, err
	}
	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
	}
	return orders, hasMore, nil
}

func (s *PostgresStore) lineItemsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64][]domain.LineItem, error) {
	rows, err := s.DB().GetLineItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("get line items: %w", err)
	}
	defer rows.Close()

	byOrder := make(map[int64][]domain.LineItem, len(orderIDs))
	for rows.Next() {
		var (
			orderID int64
			li      domain.LineItem
		)
		if err := rows.Scan(&orderID, &li.ItemID, &li.Quantity, &li.Price); err != nil {
			return nil, fmt.Errorf("scan line item: %w", err)
		}
		byOrder[orderID] = append(byOrder[orderID], li)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate line items: %w", err)
	}
	return byOrder, nil
}
