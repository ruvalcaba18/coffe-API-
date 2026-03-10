-- Create an index to speed up order lookups by creation time and ID
-- This is particularly useful for getting the latest orders or history efficiently
CREATE INDEX IF NOT EXISTS idx_orders_created_at_id ON orders (created_at DESC, id);
