ALTER TABLE products
    ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'pending';

UPDATE products
SET status = COALESCE(status, 'pending');

CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

