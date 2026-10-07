-- Stable public routes for category and brand pages, and one required
-- category for every product.
INSERT INTO categories (name, description)
SELECT 'عمومی', 'دستهٔ عمومی برای محصولات موجود'
WHERE NOT EXISTS (SELECT 1 FROM categories);

ALTER TABLE categories ADD COLUMN slug TEXT;
UPDATE categories SET slug = 'category-' || id::text WHERE slug IS NULL;
ALTER TABLE categories ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX idx_categories_slug ON categories (slug);

ALTER TABLE brands ADD COLUMN slug TEXT;
UPDATE brands SET slug = 'brand-' || id::text WHERE slug IS NULL;
ALTER TABLE brands ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX idx_brands_slug ON brands (slug);

UPDATE products
SET category_id = (SELECT id FROM categories ORDER BY id LIMIT 1)
WHERE category_id IS NULL;
ALTER TABLE products ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_category_id_fkey;
ALTER TABLE products ADD CONSTRAINT products_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT;

-- Keep previously delivered orders visible by mapping them to the final
-- supported state before removing delivered from the status constraint.
UPDATE orders SET status = 'shipped' WHERE status = 'delivered';
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('awaiting_payment', 'paid', 'processing', 'shipped', 'cancelled'));
