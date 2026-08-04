BEGIN;

CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()

);

CREATE INDEX idx_carts_user_id ON carts (user_id);

COMMIT;