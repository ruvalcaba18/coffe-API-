-- Add pickup fields to orders table
ALTER TABLE orders ADD COLUMN is_pickup BOOLEAN DEFAULT FALSE;
ALTER TABLE orders ADD COLUMN pickup_time TIMESTAMP WITH TIME ZONE;
