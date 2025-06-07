-- PayPal Integration Migration

-- Add paypal_payment_id column if it doesn't exist
ALTER TABLE IF EXISTS payments 
  ADD COLUMN IF NOT EXISTS paypal_payment_id VARCHAR(255);

-- Add paypal_order_id column if it doesn't exist
ALTER TABLE IF EXISTS payments 
  ADD COLUMN IF NOT EXISTS paypal_order_id VARCHAR(255);

-- Add paypal_customer_id column to users if it doesn't exist
ALTER TABLE IF EXISTS users 
  ADD COLUMN IF NOT EXISTS paypal_customer_id VARCHAR(255);

-- Create index on paypal_order_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_payments_paypal_order_id ON payments(paypal_order_id);

-- Create index on paypal_payment_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_payments_paypal_payment_id ON payments(paypal_payment_id); 