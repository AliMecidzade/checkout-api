package store

import (
	"checkout-api/models"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// PostgresStore is an in-memory store for items and orders.
type PostgresStore struct {
	conn *pgx.Conn
}

// NewPostgresStore creates a Store pre-loaded with seed data.
func NewPostgresStore(conn *pgx.Conn) *PostgresStore {
	s := &PostgresStore{
		conn: conn,
	}
	return s
}

func (s *PostgresStore) SignUp(ctx context.Context, password, email string) (*models.User, error) {
	var u models.User
	err := s.conn.QueryRow(ctx,
		"INSERT INTO users (password, email) VALUES ($1,$2) RETURNING id, password, email, created_at",
		password, email).
		Scan(&u.ID, &u.Password, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetItems returns all available items.
func (s *PostgresStore) GetItems(ctx context.Context) ([]*models.Item, error) {
	rows, err := s.conn.Query(ctx, "select * from items")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to run query on GetItems", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Stock, &item.CreatedAt)
		if err != nil {
			// Handle the scan error, potentially breaking the loop or logging and continuing
			fmt.Printf("unable to scan row: %v", err)
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		items = append(items, &item)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return items, nil
}

// GetItem returns a single item by ID, or nil if not found.
func (s *PostgresStore) GetItem(ctx context.Context, id int) *models.Item {
	var item models.Item

	err := s.conn.QueryRow(ctx,
		"select id,name,description,price,stock,created_at from items where id = $1", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Stock, &item.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		fmt.Printf("unable to scan row: %v", err)
		return nil
	}
	return &item
}

func (s *PostgresStore) CreateOrder(ctx context.Context, userID int, items []models.LineItem, total int, status string) (*models.Order, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderID int
	err = tx.QueryRow(ctx,
		"INSERT INTO orders (user_id, total, status) VALUES ($1,$2,$3) RETURNING id", userID, total, status).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("unable to insert order: %w", err)
	}

	for _, item := range items {
		_, err = tx.Exec(ctx,
			"INSERT INTO order_items (order_id, item_id, price, quantity) VALUES ($1,$2,$3,$4)", orderID, item.ItemID, item.Price, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("unable to insert order_item: %w", err)
		}
	}

	order := &models.Order{
		ID:     orderID,
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: status,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (s *PostgresStore) CreateUserCart(ctx context.Context, cart *models.Cart) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		fmt.Println("failed to begin transaction", err)
		return
	}
	defer tx.Rollback(ctx)

	var cartID int
	err = tx.QueryRow(ctx,
		"INSERT INTO carts (user_id) VALUES ($1) RETURNING id", cart.UserID).Scan(&cartID)

	if err != nil {
		fmt.Println("failed to insert into cart", err)
		return
	}

	for _, item := range cart.Items {
		_, err = tx.Exec(ctx, "INSERT INTO cart_items (cart_id, item_id, price, quantity) VALUES ($1,$2,$3,$4)",
			cartID, item.ItemID, item.Price, item.Quantity)
		if err != nil {
			fmt.Println("failed to insert into cart_item", err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Println("failed to commit transaction", err)
	}

}

func (s *PostgresStore) GetUserCart(ctx context.Context, userID int) *models.Cart {
	var cartID int
	err := s.conn.QueryRow(ctx,
		"SELECT id FROM carts WHERE user_id = $1", userID).Scan(&cartID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		fmt.Println("failed to load cart:", err)
		return nil
	}

	rows, err := s.conn.Query(ctx,
		"SELECT item_id, price, quantity FROM cart_items WHERE cart_id = $1 ORDER BY id", cartID)
	if err != nil {
		fmt.Println("failed to load cart items:", err)
		return nil
	}
	defer rows.Close()

	cart := &models.Cart{
		ID:     fmt.Sprintf("cart_%d", cartID),
		UserID: userID,
		Items:  []models.LineItem{},
	}
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ItemID, &li.Price, &li.Quantity); err != nil {
			fmt.Println("failed to scan cart item:", err)
			return nil
		}
		cart.Items = append(cart.Items, li)
	}
	if rows.Err() != nil {
		fmt.Println("failed to iterate cart items:", rows.Err())
		return nil
	}

	return cart
}

func (s *PostgresStore) DeleteUserCart(ctx context.Context, userID int) {
	_, err := s.conn.Exec(ctx, "DELETE FROM carts WHERE user_id = $1", userID)
	if err != nil {
		fmt.Println("failed to delete user cart", err)
	}
}

func (s *PostgresStore) UpdateCartItem(ctx context.Context, userID int, itemID int, quantity int) bool {
	if quantity <= 0 {
		return false
	}

	cmd, err := s.conn.Exec(ctx,
		"UPDATE cart_items ci SET quantity = $3 FROM carts c WHERE ci.cart_id = c.id AND c.user_id = $1 AND ci.item_id = $2",
		userID, itemID, quantity)
	if err != nil {
		fmt.Println("failed to update cart item", err)
		return false
	}
	return cmd.RowsAffected() > 0
}

func (s *PostgresStore) RemoveCartItem(ctx context.Context, userID int, itemID int) bool {
	// TODO: implement
	cmd, err := s.conn.Exec(ctx,
		"DELETE FROM cart_items ci USING carts c WHERE ci.cart_id = c.id AND c.user_id = $1 AND ci.item_id = $2",
		userID, itemID)
	if err != nil {
		fmt.Println("failed to remove cart item", err)
		return false
	}

	return cmd.RowsAffected() > 0
}
