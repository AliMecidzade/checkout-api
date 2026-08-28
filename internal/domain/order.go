package domain

import "time"

const (
	OrderStatusPending = "pending"
	OrderStatusCreated = "created"
	OrderStatusFailed  = "failed"
)

type Order struct {
	ID        int64
	UserID    int64
	Items     []LineItem
	Total     int64
	Status    string
	CreatedAt time.Time
}

type LineItem struct {
	ItemID   int64 `json:"item_id"`
	Quantity int64 `json:"quantity"`
	Price    int64 `json:"price"` // Price at time of adding
}

func (o Order) ComputeTotal() int64 {
	var sum int64
	for _, li := range o.Items {
		sum += li.Quantity * li.Price

	}
	return sum
}
