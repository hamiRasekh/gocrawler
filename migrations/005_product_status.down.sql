DROP INDEX IF EXISTS idx_products_status;
ALTER TABLE products DROP COLUMN IF EXISTS status;

