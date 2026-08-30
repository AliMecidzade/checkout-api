package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"
	"checkout-api/internal/validation"
)

type PaymentResult struct {
	Success       bool
	TransactionID string
	Error         string
}

type CreateOrderResult struct {
	Order   domain.Order
	Payment PaymentResult
}

type IdempotencyRecord struct {
	Result CreateOrderResult
	Expiry time.Time
}

type OrderService struct {
	orders repository.OrderRepo
	mu     sync.Mutex
	idem   map[string]IdempotencyRecord
}

func NewOrderService(orders repository.OrderRepo) *OrderService {
	return &OrderService{orders: orders, idem: make(map[string]IdempotencyRecord)}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []domain.LineItem, idempotencyKey string) (CreateOrderResult, error) {
	var verr validation.ValidationError
	if len(items) == 0 {
		verr.With("line_items", "required", "line_items must not be empty")
	}
	var total int64
	for i, li := range items {
		if li.Quantity <= 0 {
			verr.With(fmt.Sprintf("line_items[%d].quantity", i), "gt=0", "quantity must be greater than 0")
		}
		if li.Price < 0 {
			verr.With(fmt.Sprintf("line_items[%d].price", i), "gte=0", "price must not be negative")
		}
		total += li.Price * li.Quantity
	}
	if len(verr.Fields) != 0 {
		return CreateOrderResult{}, &verr
	}

	s.mu.Lock()
	if rec, ok := s.idem[idempotencyKey]; ok {
		if time.Now().Before(rec.Expiry) {
			s.mu.Unlock()
			return rec.Result, nil
		}
		delete(s.idem, idempotencyKey)
	}
	s.mu.Unlock()

	order, err := s.orders.CreateOrder(ctx, userID, items, total, domain.OrderStatusPending)
	if err != nil {
		return CreateOrderResult{}, fmt.Errorf("create order: %w", err)
	}

	payment := mockProcessPayment(total)

	status := domain.OrderStatusPaid
	if !payment.Success {
		status = domain.OrderStatusFailed
	}
	if err := s.orders.UpdateOrderStatus(ctx, order.ID, status); err != nil {
		return CreateOrderResult{}, fmt.Errorf("update order status: %w", err)
	}
	order.Status = status

	result := CreateOrderResult{Order: order, Payment: payment}

	s.mu.Lock()
	s.idem[idempotencyKey] = IdempotencyRecord{Result: result, Expiry: time.Now().Add(24 * time.Hour)}
	s.mu.Unlock()

	return result, nil
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, userID int64, p domain.Page) ([]domain.Order, bool, error) {
	p = normalizePage(p)

	orders, hasMore, err := s.orders.ListOrdersByUser(ctx, userID, p)
	if err != nil {
		return nil, false, fmt.Errorf("list orders: %w", err)
	}
	return orders, hasMore, nil
}

func mockProcessPayment(amount int64) PaymentResult {
	if amount > 0 && amount < 1000000 {
		return PaymentResult{
			Success:       true,
			TransactionID: fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		}
	}
	return PaymentResult{
		Success: false,
		Error:   "Payment declined",
	}
}
