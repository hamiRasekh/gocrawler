-- Create products table for embroidery designs
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    elastic_id VARCHAR(255) UNIQUE NOT NULL, -- _id from elasticsearch
    product_id VARCHAR(255), -- productId from source
    item_id VARCHAR(255), -- itemId from source
    name TEXT,
    brand VARCHAR(255),
    catalog VARCHAR(255),
    artist VARCHAR(255),
    rating DECIMAL(3, 1),
    list_price DECIMAL(10, 2),
    sale_price DECIMAL(10, 2),
    club_price DECIMAL(10, 2),
    sale_rank INTEGER,
    customer_interest_index INTEGER,
    in_stock BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    is_buyable BOOLEAN DEFAULT true,
    licensed BOOLEAN DEFAULT false,
    is_applique BOOLEAN DEFAULT false,
    is_cross_stitch BOOLEAN DEFAULT false,
    is_pdf_available BOOLEAN DEFAULT false,
    is_fsl BOOLEAN DEFAULT false,
    is_heat_transfer BOOLEAN DEFAULT false,
    is_design_used_in_project BOOLEAN DEFAULT false,
    in_custom_pack BOOLEAN DEFAULT false,
    definition_name VARCHAR(255),
    product_type VARCHAR(255),
    gtin VARCHAR(255),
    color_sequence TEXT,
    design_keywords TEXT,
    categories TEXT, -- comma separated
    categories_list TEXT, -- JSON array
    keywords TEXT, -- JSON array
    sales TEXT,
    sales_list TEXT, -- JSON array
    sale_end_date TIMESTAMP,
    year_created TIMESTAMP,
    applied_discount_id INTEGER,
    is_multiple_variants_available BOOLEAN DEFAULT false,
    variants TEXT, -- JSON array of variants
    raw_data TEXT, -- Full JSON from API
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_products_elastic_id ON products(elastic_id);
CREATE INDEX IF NOT EXISTS idx_products_product_id ON products(product_id);
CREATE INDEX IF NOT EXISTS idx_products_item_id ON products(item_id);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_catalog ON products(catalog);
CREATE INDEX IF NOT EXISTS idx_products_in_stock ON products(in_stock);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products(updated_at);
CREATE INDEX IF NOT EXISTS idx_products_sale_rank ON products(sale_rank);
CREATE INDEX IF NOT EXISTS idx_products_rating ON products(rating);

