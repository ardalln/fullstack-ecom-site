ALTER TABLE users ADD COLUMN username TEXT;
UPDATE users SET username = 'user_' || id::text WHERE username IS NULL OR btrim(username) = '';
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
CREATE UNIQUE INDEX users_username_lower_key ON users (lower(username));

ALTER TABLE products ADD COLUMN is_popular BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE banners (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    subtitle TEXT NOT NULL DEFAULT '',
    desktop_image_url TEXT NOT NULL,
    mobile_image_url TEXT NOT NULL,
    link_url TEXT NOT NULL DEFAULT '',
    button_label TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shipping_methods (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    province TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    price BIGINT NOT NULL CHECK (price >= 0),
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((province = '' AND city = '') OR province <> '')
);
CREATE INDEX idx_shipping_methods_active_location ON shipping_methods (is_active, province, city, sort_order);
INSERT INTO shipping_methods (name, description, province, city, price, is_active, sort_order)
VALUES ('ارسال سراسری', 'روش ارسال پیش‌فرض فروشگاه', '', '', 0, true, 0);

CREATE TABLE coupons (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('percent', 'fixed')),
    value BIGINT NOT NULL CHECK (value > 0),
    minimum_subtotal BIGINT NOT NULL DEFAULT 0 CHECK (minimum_subtotal >= 0),
    maximum_discount BIGINT NULL CHECK (maximum_discount IS NULL OR maximum_discount > 0),
    usage_limit BIGINT NULL CHECK (usage_limit IS NULL OR usage_limit > 0),
    used_count BIGINT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    starts_at TIMESTAMPTZ NULL,
    ends_at TIMESTAMPTZ NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at),
    CHECK (kind <> 'percent' OR value <= 100)
);
CREATE UNIQUE INDEX idx_coupons_code_lower ON coupons (lower(code));

ALTER TABLE orders
    ADD COLUMN items_amount BIGINT NOT NULL DEFAULT 0 CHECK (items_amount >= 0),
    ADD COLUMN discount_amount BIGINT NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    ADD COLUMN shipping_cost BIGINT NOT NULL DEFAULT 0 CHECK (shipping_cost >= 0),
    ADD COLUMN coupon_id BIGINT NULL REFERENCES coupons(id) ON DELETE SET NULL,
    ADD COLUMN coupon_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN shipping_method_id BIGINT NULL REFERENCES shipping_methods(id) ON DELETE SET NULL,
    ADD COLUMN shipping_method_name TEXT NOT NULL DEFAULT 'ارسال سراسری',
    ADD COLUMN tracking_code TEXT NOT NULL DEFAULT '';
UPDATE orders SET items_amount = total_amount WHERE items_amount = 0 AND total_amount > 0;
UPDATE orders SET tracking_code = 'TRK-' || upper(substr(md5(id::text || created_at::text), 1, 12)) WHERE tracking_code = '';
CREATE UNIQUE INDEX idx_orders_tracking_code ON orders (tracking_code);
