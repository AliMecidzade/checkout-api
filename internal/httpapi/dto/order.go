package dto

import "time"

type CreateOrderRequest struct {
	LineItems []LineItemRequest `json:"line_items"`
	Total     int64             `json:"total"`
}

type LineItemRequest struct {
	ItemID   int64 `json:"item_id"`
	Quantity int64 `json:"quantity"`
	Price    int64 `json:"price"`
}

type PaymentResult struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type CreateOrderResponse struct {
	Order   OrderResponse `json:"order"`
	Payment PaymentResult `json:"payment"`
}

type LineItemResponse struct {
	ItemID   int64 `json:"item_id"`
	Quantity int64 `json:"quantity"`
	Price    int64 `json:"price"`
}

type OrderResponse struct {
	ID        int64              `json:"id"`
	UserID    int64              `json:"user_id"`
	Items     []LineItemResponse `json:"line_items"`
	Total     int64              `json:"total"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
