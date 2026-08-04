BEGIN;

DELETE FROM users WHERE email IN ('alice@example.com', 'bob@example.com');

COMMIT;