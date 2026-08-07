BEGIN;

INSERT INTO line_items 
SELECT * FROM order_items;

-- Copy carries explicit id values, so the SERIAL sequence was not advanced.
-- Without this, the next INSERT using nextval('line_items_id_seq') collides
-- with existing ids and fails with duplicate key on line_items_pkey.
SELECT setval('line_items_id_seq', (SELECT COALESCE(max(id), 1) FROM line_items));

COMMIT;
