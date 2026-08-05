BEGIN;
CREATE INDEX idx_orders_user_status ON orders (user_id, status);
COMMIT;