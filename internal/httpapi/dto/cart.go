package dto

import "time"

type UpsertCartItemRequest struct {
	Quantity int64 `json:"quantity"`
}

type CartItemResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Stock       int64     `json:"stock"`
	Quantity    int64     `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
}

type CartResponse struct {
	UserID int64              `json:"user_id"`
	Items  []CartItemResponse `json:"items"`
}
