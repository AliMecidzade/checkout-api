package domain

import "time"

type CartItem struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	Stock       int64
	Quantity    int64
	CreatedAt   time.Time
}


type Cart struct {
	UserID   int64
	ItemID   int64
	Quantity int64
}

