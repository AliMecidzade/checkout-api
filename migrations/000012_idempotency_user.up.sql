BEGIN;

-- Scope idempotency keys per user so a key issued by one user
-- cannot be replayed by another user.
-- Clean out any cached records first: they predate user scoping and
-- have no user_id to backfill.
DELETE FROM idempotency;

ALTER TABLE idempotency ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE idempotency DROP CONSTRAINT idempotency_pkey;
ALTER TABLE idempotency ADD PRIMARY KEY (user_id, key);

CREATE INDEX idx_idempotency_key ON idempotency (key);

COMMIT;