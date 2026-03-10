-- ==========================================================
-- 1. USERS PERFORMANCE & DENORMALIZATION
-- ==========================================================

-- Standardize indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

-- Denormalization: Store total metrics to avoid heavy aggregations on the dashboard/profile
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_orders_completed INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_spent DECIMAL(12, 2) DEFAULT 0.0;

-- Trigger to update user stats when an order is completed
CREATE OR REPLACE FUNCTION update_user_order_stats() 
RETURNS TRIGGER AS $$
BEGIN
    -- Only update if the order status changes to 'Completed' (or similar success state)
    -- Assuming 'Completed' is the final state. If it was 'Pending' and now 'Completed':
    IF (NEW.status = 'Completed' AND (OLD.status IS NULL OR OLD.status != 'Completed')) THEN
        UPDATE users 
        SET total_orders_completed = total_orders_completed + 1,
            total_spent = total_spent + NEW.total
        WHERE id = NEW.user_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_user_order_stats ON orders;
CREATE TRIGGER trg_user_order_stats
AFTER UPDATE ON orders
FOR EACH ROW EXECUTE FUNCTION update_user_order_stats();

-- ==========================================================
-- 2. ORDERS PERFORMANCE & INTEGRITY
-- ==========================================================

-- Indexes for filtering by status and searching by user
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- Denormalization: Store quantity of items to avoid joining order_items just for a count
ALTER TABLE orders ADD COLUMN IF NOT EXISTS items_count INT DEFAULT 0;

-- Trigger to keep items_count in sync
CREATE OR REPLACE FUNCTION update_order_items_count() 
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE orders SET items_count = items_count + NEW.quantity WHERE id = NEW.order_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE orders SET items_count = items_count - OLD.quantity WHERE id = OLD.order_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        UPDATE orders SET items_count = items_count - OLD.quantity + NEW.quantity WHERE id = NEW.order_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_order_items_count ON order_items;
CREATE TRIGGER trg_order_items_count
AFTER INSERT OR UPDATE OR DELETE ON order_items
FOR EACH ROW EXECUTE FUNCTION update_order_items_count();

-- ==========================================================
-- 3. COUPONS INTEGRITY
-- ==========================================================

-- Index for expiration checks (used when listing or validating)
CREATE INDEX IF NOT EXISTS idx_coupons_active_dates ON coupons (is_active, start_date, end_date);

-- Trigger to automatically increment used_count when an order with a coupon is created
CREATE OR REPLACE FUNCTION increment_coupon_usage() 
RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.coupon_code IS NOT NULL AND NEW.coupon_code != '') THEN
        UPDATE coupons SET used_count = used_count + 1 WHERE code = NEW.coupon_code;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_increment_coupon_usage ON orders;
CREATE TRIGGER trg_increment_coupon_usage
AFTER INSERT ON orders
FOR EACH ROW EXECUTE FUNCTION increment_coupon_usage();

-- ==========================================================
-- 4. INITIAL SYNC
-- ==========================================================

-- Sync user stats
UPDATE users u
SET total_orders_completed = (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id AND o.status = 'Completed'),
    total_spent = (SELECT COALESCE(SUM(total), 0) FROM orders o WHERE o.user_id = u.id AND o.status = 'Completed');

-- Sync order item counts
UPDATE orders o
SET items_count = (SELECT COALESCE(SUM(quantity), 0) FROM order_items oi WHERE oi.order_id = o.id);
