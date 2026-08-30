package postgres

import (
	"context"
	"errors"
	"fmt"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) CreateUser(ctx context.Context, email string, hash []byte) error {
	if _, err := s.DB().InsertUser(ctx, email, hash); err != nil {
		if pgCode(err) == "23505" {
			return repository.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := s.DB().GetUserByEmail(ctx, email).Scan(&u.ID, &u.Email, &u.Hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}
