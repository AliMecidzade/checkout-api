BEGIN;

CREATE TABLE idempotency (
    key        TEXT PRIMARY KEY,
    response   BYTEA,
    status     INTEGER NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

COMMIT;
