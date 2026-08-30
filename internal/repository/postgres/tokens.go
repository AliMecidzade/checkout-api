package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"checkout-api/internal/domain"
	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) SaveRefreshToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	if _, err := s.DB().InsertRefreshToken(ctx, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (s *PostgresStore) FindRefreshToken(ctx context.Context, tokenHash []byte) (domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := s.DB().FindRefreshToken(ctx, tokenHash).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.IsActive, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RefreshToken{}, repository.ErrNotFound
	}
	if err != nil {
		return domain.RefreshToken{}, fmt.Errorf("find refresh token: %w", err)
	}
	return rt, nil
}

func (s *PostgresStore) DeactivateRefreshToken(ctx context.Context, tokenHash []byte) error {
	if _, err := s.DB().DeactivateRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("deactivate refresh token: %w", err)
	}
	return nil
}

func (s *PostgresStore) RotateRefreshToken(ctx context.Context, oldHash []byte, newHash []byte, userID int64, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.WithTx(tx)

	if _, err := q.DeactivateRefreshToken(ctx, oldHash); err != nil {
		return fmt.Errorf("deactivate old refresh token: %w", err)
	}
	if _, err := q.InsertRefreshToken(ctx, userID, newHash, expiresAt); err != nil {
		return fmt.Errorf("insert new refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, tokenHash []byte) error {
	return s.DeactivateRefreshToken(ctx, tokenHash)
}
