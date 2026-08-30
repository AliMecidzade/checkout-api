package dto

import "time"

type ItemResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Stock       int64     `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
}
