-- Richer catalog: hierarchical categories, brands, searchable product slugs,
-- image galleries and structured key/value attributes.
CREATE TABLE brands (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_brands_name_lower ON brands (lower(name));

INSERT INTO brands (name, description) VALUES ('عمومی', 'برند عمومی برای محصولات موجود')
ON CONFLICT DO NOTHING;

ALTER TABLE categories
    ADD COLUMN parent_id BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    ADD CONSTRAINT categories_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id);
CREATE INDEX idx_categories_parent_id ON categories (parent_id, name);

ALTER TABLE products
    ADD COLUMN slug TEXT,
    ADD COLUMN image_urls TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN brand_id BIGINT REFERENCES brands (id) ON DELETE RESTRICT;

UPDATE products
SET slug = CASE name
        WHEN 'کیبورد مکانیکی' THEN 'mechanical-keyboard-' || id::text
        WHEN 'ماوس بی‌سیم' THEN 'wireless-mouse-' || id::text
        WHEN 'هاب USB-C' THEN 'usb-c-hub-' || id::text
        ELSE 'product-' || id::text
    END,
    image_urls = CASE WHEN image_url <> '' THEN ARRAY[image_url]::text[] ELSE '{}'::text[] END,
    brand_id = (SELECT id FROM brands WHERE name = 'عمومی' ORDER BY id LIMIT 1);

ALTER TABLE products ALTER COLUMN slug SET NOT NULL;
ALTER TABLE products ALTER COLUMN brand_id SET NOT NULL;
CREATE UNIQUE INDEX idx_products_slug ON products (slug);
CREATE INDEX idx_products_brand_id ON products (brand_id, id DESC);
CREATE INDEX idx_products_name_lower ON products (lower(name));
