BEGIN;

INSERT INTO users (password, email) VALUES
    ('alice-password', 'alice@example.com'),
    ('bob-password',   'bob@example.com');

COMMIT;