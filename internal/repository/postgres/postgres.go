package postgres

import (
	"errors"

	"checkout-api/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) DB() *Query {
	return &Query{DBTX: s.pool}
}

func (s *PostgresStore) WithTx(tx pgx.Tx) *Query {
	return &Query{DBTX: tx}
}

var (
	_ repository.CatalogRepo = (*PostgresStore)(nil)
	_ repository.OrderRepo   = (*PostgresStore)(nil)
	_ repository.CartRepo    = (*PostgresStore)(nil)
	_ repository.UserRepo    = (*PostgresStore)(nil)
	_ repository.TokenRepo   = (*PostgresStore)(nil)
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
