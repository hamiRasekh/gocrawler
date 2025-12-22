-- Drop indexes
DROP INDEX IF EXISTS idx_products_rating;
DROP INDEX IF EXISTS idx_products_sale_rank;
DROP INDEX IF EXISTS idx_products_updated_at;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_in_stock;
DROP INDEX IF EXISTS idx_products_catalog;
DROP INDEX IF EXISTS idx_products_brand;
DROP INDEX IF EXISTS idx_products_item_id;
DROP INDEX IF EXISTS idx_products_product_id;
DROP INDEX IF EXISTS idx_products_elastic_id;

-- Drop products table
DROP TABLE IF EXISTS products;

