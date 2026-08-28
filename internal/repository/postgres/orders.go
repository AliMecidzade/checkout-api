package postgres

import (
	"context"
	"errors"
	"fmt"

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
	if err := q.InsertOrderReturning(ctx, userID, total, status).Scan(&orderID); err != nil {
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
		ID:     orderID,
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: status,
	}, nil
}

func (s *PostgresStore) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	if _, err := s.DB().UpdateOrderStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("update order %d status: %w", orderID, err)
	}
	return nil
}

func (s *PostgresStore) GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	rows, err := s.DB().GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get orders for user %d: %w", userID, err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	orderIDs := make([]int64, 0)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
		orderIDs = append(orderIDs, o.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	if len(orderIDs) == 0 {
		return orders, nil
	}

	itemsByOrder, err := s.lineItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
	}
	return orders, nil
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
