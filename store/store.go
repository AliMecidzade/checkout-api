package store

import (
	"checkout-api/models"
	"fmt"
)

// Store is an in-memory store for items and orders.
type Store struct {
	items       map[int]*models.Item
	orders      map[int]*models.Order
	nextOrderID int
	carts       map[int]*models.Cart
}

// NewStore creates a Store pre-loaded with seed data.
func NewStore() *Store {
	s := &Store{
		items:       make(map[int]*models.Item),
		orders:      make(map[int]*models.Order),
		nextOrderID: 1,
		carts:       make(map[int]*models.Cart),
	}

	s.items[1] = &models.Item{ID: 1, Name: "Laptop", Description: "A fast laptop", Price: 120000, Stock: 10}
	s.items[2] = &models.Item{ID: 2, Name: "Mouse", Description: "Wireless mouse", Price: 2500, Stock: 50}
	s.items[3] = &models.Item{ID: 3, Name: "Keyboard", Description: "Mechanical keyboard", Price: 8000, Stock: 30}

	return s
}

// GetItems returns all available items.
func (s *Store) GetItems() []*models.Item {
	items := make([]*models.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

// GetItem returns a single item by ID, or nil if not found.
func (s *Store) GetItem(id int) *models.Item {
	return s.items[id]
}

// CreateOrder creates a new order and returns it.
func (s *Store) CreateOrder(userID int, items []models.LineItem, total int, status string) *models.Order {
	order := &models.Order{
		ID:     s.nextOrderID,
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: status,
	}
	s.orders[order.ID] = order
	s.nextOrderID++
	return order
}

func (s *Store) CreateUserCart(cart *models.Cart) {
	s.carts[cart.UserID] = cart
}

func (s *Store) GetUserCart(userID int) *models.Cart {
	if c, ok := s.carts[userID]; ok {
		return c
	}
	return nil
}

func (s *Store) UpdateCartItemQuantity(userID int, itemId int, quantity int) error {
	cart, ok := s.carts[userID]
	if !ok {
		return fmt.Errorf("cart not found")
	}

	for i, item := range cart.Items {
		if item.ItemID == itemId {
			cart.Items[i].Quantity = quantity
			return nil
		}

	}

	return fmt.Errorf("item %d not found in cart", itemId)

}

func (s *Store) DeleteItemFromCart(userID int, itemId int) error {
	cart, ok := s.carts[userID]
	if !ok {
		return fmt.Errorf("cart not found")
	}

	for i, item := range cart.Items {
		if item.ItemID == itemId {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			if len(cart.Items) == 0 {
				s.DeleteCart(userID)
			}
			return nil
		}
	}

	return fmt.Errorf("item %d not found in cart", itemId)
}

func (s *Store) DeleteCart(userID int) error {
	if _, ok := s.carts[userID]; !ok {
		return fmt.Errorf("cart not found")
	}
	delete(s.carts, userID)
	return nil
}
