-- Catalog management and administrative account controls.
CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_categories_name_lower ON categories (lower(name));

INSERT INTO categories (name, description) VALUES ('عمومی', 'محصولات بدون دسته‌بندی مشخص')
ON CONFLICT DO NOTHING;

-- Translate only the three original demo rows when upgrading an older checkout.
UPDATE products SET name = 'کیبورد مکانیکی', description = 'کیبورد مکانیکی ۷۵ درصد با کلیدهای قابل تعویض'
WHERE name = 'Mechanical Keyboard' AND description = 'Hot-swappable 75% mechanical keyboard';
UPDATE products SET name = 'ماوس بی‌سیم', description = 'ماوس ارگونومیک بی‌سیم ۲٫۴ گیگاهرتز'
WHERE name = 'Wireless Mouse' AND description = 'Ergonomic 2.4GHz wireless mouse';
UPDATE products SET name = 'هاب USB-C', description = 'هاب هفت‌کاره USB-C با درگاه HDMI'
WHERE name = 'USB-C Hub' AND description = '7-in-1 USB-C hub with HDMI';

ALTER TABLE products
    ADD COLUMN category_id BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    ADD COLUMN image_url TEXT NOT NULL DEFAULT '';

UPDATE products
SET category_id = (SELECT id FROM categories ORDER BY id LIMIT 1)
WHERE category_id IS NULL;

CREATE INDEX idx_products_category_id ON products (category_id, id DESC);

ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
CREATE INDEX idx_users_created_at ON users (created_at DESC, id DESC);

ALTER TABLE orders ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('awaiting_payment', 'paid', 'processing', 'shipped', 'delivered', 'cancelled'));
CREATE INDEX idx_orders_status_created_at ON orders (status, created_at DESC, id DESC);
