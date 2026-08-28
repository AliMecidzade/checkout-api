package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Query struct {
	DBTX DBTX
}

func (q *Query) GetItemsOffset(ctx context.Context, limit int64, offset int64) (pgx.Rows, error) {
	return q.DBTX.Query(ctx,
		"select id, name, description, price, stock, created_at from items order by created_at desc, id desc limit $1 offset $2",
		limit, offset,
	)
}

func (q *Query) GetItemsAfter(ctx context.Context, limit int64, cursorCreatedAt time.Time, cursorID int64) (pgx.Rows, error) {
	return q.DBTX.Query(ctx,
		"select id, name, description, price, stock, created_at from items where (created_at, id) < ($2, $3) order by created_at desc, id desc limit $1",
		limit, cursorCreatedAt, cursorID,
	)
}

func (q *Query) GetItemByID(ctx context.Context, id int64) pgx.Row {
	return q.DBTX.QueryRow(ctx, "select id, name, description, price, stock, created_at from items where id = $1", id)
}

func (q *Query) GetItemByIDForUpdate(ctx context.Context, id int64) pgx.Row {
	return q.DBTX.QueryRow(ctx, "select id, name, description, price, stock, created_at from items where id = $1 for update", id)
}

func (q *Query) GetItemsFromUserCart(ctx context.Context, userID int64) (pgx.Rows, error) {
	return q.DBTX.Query(ctx,
		`select i.id, i.name, i.description, i.price, i.stock, i.created_at, c.quantity
		from carts c
		inner join items i on i.id = c.item_id
		where c.user_id = $1`,
		userID,
	)
}

func (q *Query) UpsertCart(ctx context.Context, userID int64, itemID int64, quantity int64) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx,
		"insert into carts (user_id, item_id, quantity) values ($1, $2, $3) on conflict (user_id, item_id) do update set quantity = excluded.quantity",
		userID, itemID, quantity,
	)
}

func (q *Query) DeleteCartByUserID(ctx context.Context, userID int64) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "delete from carts where user_id = $1", userID)
}

func (q *Query) DeleteItemFromUserCart(ctx context.Context, userID int64, itemID int64) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "delete from carts where user_id = $1 and item_id = $2", userID, itemID)
}

func (q *Query) GetUserOrders(ctx context.Context, userID int64) (pgx.Rows, error) {
	return q.DBTX.Query(ctx,
		"select id, user_id, total, status, created_at from orders where user_id = $1 order by created_at desc, id desc",
		userID,
	)
}

func (q *Query) GetLineItemsByOrderIDs(ctx context.Context, orderIDs []int64) (pgx.Rows, error) {
	return q.DBTX.Query(ctx,
		"select order_id, item_id, quantity, price from line_items where order_id = any($1) order by id",
		orderIDs,
	)
}

func (q *Query) InsertOrderReturning(ctx context.Context, userID int64, total int64, status string) pgx.Row {
	return q.DBTX.QueryRow(ctx, "insert into orders (user_id, total, status) values ($1, $2, $3) returning id", userID, total, status)
}

func (q *Query) UpdateOrderStatus(ctx context.Context, orderID int64, status string) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "update orders set status = $1 where id = $2", status, orderID)
}

func (q *Query) InsertLineItem(ctx context.Context, orderID int64, itemID int64, price int64, quantity int64) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "insert into line_items (order_id, item_id, price, quantity) values ($1, $2, $3, $4)", orderID, itemID, price, quantity)
}

func (q *Query) DecrementItemStock(ctx context.Context, itemID int64, quantity int64) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "update items set stock = stock - $1 where id = $2", quantity, itemID)
}

func (q *Query) InsertUser(ctx context.Context, email string, hash []byte) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "insert into users (email, hash) values ($1, $2)", email, hash)
}

func (q *Query) GetUserByEmail(ctx context.Context, email string) pgx.Row {
	return q.DBTX.QueryRow(ctx, "select id, email, hash from users where email = $1", email)
}

func (q *Query) InsertRefreshToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "insert into refresh_tokens (user_id, token_hash, is_active, expires_at) values ($1, $2, true, $3)", userID, tokenHash, expiresAt)
}

func (q *Query) FindRefreshToken(ctx context.Context, tokenHash []byte) pgx.Row {
	return q.DBTX.QueryRow(ctx, "select id, user_id, token_hash, is_active, expires_at, created_at, revoked_at from refresh_tokens where token_hash = $1", tokenHash)
}

func (q *Query) DeactivateRefreshToken(ctx context.Context, tokenHash []byte) (pgconn.CommandTag, error) {
	return q.DBTX.Exec(ctx, "update refresh_tokens set is_active = false, revoked_at = now() where token_hash = $1", tokenHash)
}
