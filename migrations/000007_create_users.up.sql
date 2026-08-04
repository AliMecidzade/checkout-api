BEGIN;

CREATE TABLE users (
    id         SERIAL  PRIMARY KEY,
    password   TEXT    NOT NULL,
    email      TEXT    NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMIT;