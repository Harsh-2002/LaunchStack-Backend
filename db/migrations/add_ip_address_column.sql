-- Add IP address column to instances table
ALTER TABLE instances ADD COLUMN IF NOT EXISTS ip_address VARCHAR(50); 