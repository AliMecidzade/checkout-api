package models

// Define your domain models here:
// - Item: ID (int), Name (string), Description (string), Price (int, in cents), Stock (int)
// - LineItem: ItemID (int), Quantity (int), Price (int)
// - Order: ID (int), UserID (int), Items ([]LineItem), Total (int), Status (string)

type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	StockSize   int    `json:"stock_size"`

}
type LineItem struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
	Price    int `json:"price"`
}
type Order struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Items      []LineItem `json:"line_items"`
	TotalPrice int        `json:"total_price"`
	Status     string     `json:"status"`
}

//type Cart struct