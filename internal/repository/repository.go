package repository

import (
	"checkout-api/internal/domain"
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrInsufficient = errors.New("insufficient stock")
)

type CatalogRepo interface {
	ListItems(ctx context.Context, p domain.Page) ([]domain.Item, bool, error)
	GetItem(ctx context.Context, id int64) (domain.Item, error)
}

type OrderRepo interface {
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	CreateOrder(ctx context.Context, userID int64, items []domain.LineItem, total int64, status string) (domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]domain.Order, error)
}

type CartRepo interface {
	UpsertCartItem(ctx context.Context, userID int64, itemID int64, quantity int64) error
	GetUserCart(ctx context.Context, userID int64) ([]domain.CartItem, error)
	DeleteUserCart(ctx context.Context, userID int64) error
	RemoveCartItem(ctx context.Context, userID int64, itemID int64) error
}

type UserRepo interface {
	CreateUser(ctx context.Context, email string, hash []byte) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type TokenRepo interface {
	FindRefreshToken(ctx context.Context, tokenHash []byte) (domain.RefreshToken, error)
	DeactivateRefreshToken(ctx context.Context, tokenHash []byte) error

	RotateRefreshToken(ctx context.Context, oldHash []byte, newHash []byte, userID int64, expiresAt time.Time) error

	SaveRefreshToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error

	RevokeRefreshToken(ctx context.Context, tokenHash []byte) error
}
