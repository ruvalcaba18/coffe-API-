-- 1. Indexes for fast filtering and sorting
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
CREATE INDEX IF NOT EXISTS idx_products_price ON products (price);
CREATE INDEX IF NOT EXISTS idx_products_name ON products (name);

-- 2. Cascading deletes for referential integrity
-- Ensure that deleting a product cleans up related transient data (reviews/favorites)
-- Note: order_items should usually stay for historical records, but if the user wants "light" DB:
ALTER TABLE order_items 
DROP CONSTRAINT IF EXISTS order_items_product_id_fkey,
ADD CONSTRAINT order_items_product_id_fkey 
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

-- 3. Optimization for review lookups
CREATE INDEX IF NOT EXISTS idx_reviews_product_rating ON reviews (product_id, rating);

-- 4. Denormalization for high-performance reads (Optional but recommended for "light" queries)
-- Adding columns to avoid heavy JOINs during listing
ALTER TABLE products ADD COLUMN IF NOT EXISTS average_rating DECIMAL(3,2) DEFAULT 0.0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS review_count INT DEFAULT 0;

-- 5. Trigger to automatically keep ratings in sync (Best Practice for performance)
CREATE OR REPLACE FUNCTION update_product_rating() 
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        UPDATE products 
        SET average_rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE product_id = NEW.product_id),
            review_count = (SELECT COUNT(*) FROM reviews WHERE product_id = NEW.product_id)
        WHERE id = NEW.product_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE products 
        SET average_rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE product_id = OLD.product_id),
            review_count = (SELECT COUNT(*) FROM reviews WHERE product_id = OLD.product_id)
        WHERE id = OLD.product_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_rating ON reviews;
CREATE TRIGGER trg_update_rating
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW EXECUTE FUNCTION update_product_rating();

-- Initial sync for existing data
UPDATE products p
SET average_rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews r WHERE r.product_id = p.id),
    review_count = (SELECT COUNT(*) FROM reviews r WHERE r.product_id = p.id);
