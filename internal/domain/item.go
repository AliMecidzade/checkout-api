package domain

import "time"

type Item struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Stock       int64
	CreatedAt   time.Time
}
