BEGIN;

ALTER TABLE items ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();

CREATE OR REPLACE FUNCTION is_item_in_stock(p_item_id INTEGER, p_qty INTEGER)
RETURNS BOOLEAN
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN EXISTS (SELECT 1 FROM items WHERE id = p_item_id AND stock >= p_qty);
END;
$$;

CREATE OR REPLACE PROCEDURE increase_item_stock(p_item_id INTEGER, p_qty INTEGER)
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE items SET stock = stock + p_qty WHERE id = p_item_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'item % not found', p_item_id;
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS items_set_updated_at ON items;

CREATE TRIGGER items_set_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

COMMIT;
