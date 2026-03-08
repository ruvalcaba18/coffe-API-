-- Add pickup_location field to orders table to allow users to choose where to collect their order
ALTER TABLE orders ADD COLUMN pickup_location VARCHAR(255);
