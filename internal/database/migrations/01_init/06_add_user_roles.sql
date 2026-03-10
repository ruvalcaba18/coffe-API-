-- Add role column to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'customer';

-- Update a specific user to be admin for testing (optional but helpful)
-- UPDATE users SET role = 'admin' WHERE email = 'admin@coffee.com';
