BEGIN;

-- Remove duplicate keys before restoring the global key primary key.
-- Keep the earliest (lowest user_id) record for each key.
DELETE FROM idempotency a
USING idempotency b
WHERE a.key = b.key
  AND a.user_id > b.user_id;

DROP INDEX IF EXISTS idx_idempotency_key;

ALTER TABLE idempotency DROP CONSTRAINT idempotency_pkey;
ALTER TABLE idempotency ADD PRIMARY KEY (key);

ALTER TABLE idempotency DROP COLUMN user_id;

COMMIT;