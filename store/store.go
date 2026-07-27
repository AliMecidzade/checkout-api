package store

import "checkout-api/models"

// Implement in-memory Store with NewStore() constructor.
// Seed data: Laptop (120000), Mouse (2500), Keyboard (8000) — prices in cents.
// Methods: GetItems(), GetItem(id int), CreateOrder(userID int, items []models.LineItem, total int, status string)

type Store struct {
	Items       map[int]*models.Item
	Orders      map[int]*models.Order
	NextOrderId int
}

func NewStore() *Store {
	store := &Store{
		Items:       make(map[int]*models.Item),
		Orders:      make(map[int]*models.Order),
		NextOrderId: 1,
	}

	store.Items[1] = &models.Item{
		ID:          1,
		Name:        "Laptop",
		Description: "Laptop",
		Price:       120000,
		StockSize:   100,
	}
	store.Items[2] = &models.Item{
		ID:          2,
		Name:        "Mouse",
		Description: "Mouse",
		Price:       2500,
		StockSize:   100,
	}
	store.Items[3] = &models.Item{
		ID:          3,
		Name:        "Keyboard",
		Description: "Keyboard",
		Price:       8000,
		StockSize:   100,
	}

	return store
}

func (store *Store) GetItems() []*models.Item {
	items := make([]*models.Item, 0, len(store.Items))
	for _, item := range store.Items {
		items = append(items, item)
	}
	return items
}
func (store *Store) GetItem(id int) *models.Item {
	return store.Items[id]
}
func (store *Store) CreateOrder(userID int,
	items []models.LineItem,
	total int,
	status string) *models.Order {
	return nil
}
