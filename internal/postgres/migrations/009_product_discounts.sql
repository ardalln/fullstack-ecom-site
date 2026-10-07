ALTER TABLE products
    ADD COLUMN discount_price BIGINT NOT NULL DEFAULT 0
    CHECK (discount_price >= 0 AND (discount_price = 0 OR discount_price < price));

INSERT INTO categories (name, description, slug)
SELECT 'کتونی', 'مدل‌های ورزشی و روزمره برای هر مسیر', 'sneakers'
WHERE NOT EXISTS (SELECT 1 FROM categories WHERE slug = 'sneakers');

INSERT INTO categories (name, description, slug, parent_id)
SELECT 'رانینگ', 'کتونی‌های سبک برای دویدن و تمرین', 'running', parent.id
FROM categories parent
WHERE parent.slug = 'sneakers'
  AND NOT EXISTS (SELECT 1 FROM categories WHERE slug = 'running');

INSERT INTO categories (name, description, slug, parent_id)
SELECT 'روزمره', 'مدل‌های راحت برای استفاده روزانه', 'everyday', parent.id
FROM categories parent
WHERE parent.slug = 'sneakers'
  AND NOT EXISTS (SELECT 1 FROM categories WHERE slug = 'everyday');

INSERT INTO brands (name, description, slug)
SELECT 'Nike', 'مدل‌های ورزشی و خیابانی نایکی', 'nike'
WHERE NOT EXISTS (SELECT 1 FROM brands WHERE slug = 'nike');

INSERT INTO brands (name, description, slug)
SELECT 'Adidas', 'کالکشن‌های ورزشی و روزمره آدیداس', 'adidas'
WHERE NOT EXISTS (SELECT 1 FROM brands WHERE slug = 'adidas');

INSERT INTO brands (name, description, slug)
SELECT 'New Balance', 'کتونی‌های راحت و کلاسیک نیوبالانس', 'new-balance'
WHERE NOT EXISTS (SELECT 1 FROM brands WHERE slug = 'new-balance');

-- Replace only the untouched starter catalog so a fresh installation opens as
-- a sneaker store. Existing products and customer-created catalog data remain intact.
UPDATE products SET
    name = 'کتونی رانینگ نایکی مدل ایر',
    slug = 'nike-air-runner-' || id::text,
    description = 'کتونی سبک و خوش‌فرم برای دویدن، تمرین و استفاده روزانه.',
    price = 6890000,
    discount_price = 5790000,
    image_url = 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=1000&q=85',
    image_urls = ARRAY['https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=1000&q=85'],
    category_id = (SELECT id FROM categories WHERE slug = 'running'),
    brand_id = (SELECT id FROM brands WHERE slug = 'nike')
WHERE name = 'کیبورد مکانیکی';

UPDATE products SET
    name = 'کتونی روزمره آدیداس مدل سامبا',
    slug = 'adidas-samba-' || id::text,
    description = 'طراحی کلاسیک و راحت برای استایل روزمره و قدم‌زدن‌های طولانی.',
    price = 6290000,
    discount_price = 0,
    image_url = 'https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=1000&q=85',
    image_urls = ARRAY['https://images.unsplash.com/photo-1552346154-21d32810aba3?auto=format&fit=crop&w=1000&q=85'],
    category_id = (SELECT id FROM categories WHERE slug = 'everyday'),
    brand_id = (SELECT id FROM brands WHERE slug = 'adidas')
WHERE name = 'ماوس بی‌سیم';

UPDATE products SET
    name = 'کتونی نیوبالانس مدل ۵۳۰',
    slug = 'new-balance-530-' || id::text,
    description = 'راحتی روزانه با فرم کلاسیک و ترکیب رنگ‌های ساده و کاربردی.',
    price = 7490000,
    discount_price = 6790000,
    image_url = 'https://images.unsplash.com/photo-1600185365483-26d7a4cc7519?auto=format&fit=crop&w=1000&q=85',
    image_urls = ARRAY['https://images.unsplash.com/photo-1600185365483-26d7a4cc7519?auto=format&fit=crop&w=1000&q=85'],
    category_id = (SELECT id FROM categories WHERE slug = 'everyday'),
    brand_id = (SELECT id FROM brands WHERE slug = 'new-balance')
WHERE name = 'هاب USB-C';
