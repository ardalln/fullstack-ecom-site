ALTER TABLE products ADD COLUMN variants JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE order_items ADD COLUMN variant_name TEXT NOT NULL DEFAULT '';

CREATE TABLE product_comments (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 2000),
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_comments_visibility ON product_comments (product_id, is_approved, created_at DESC);
CREATE INDEX idx_product_comments_admin ON product_comments (is_approved, created_at DESC);
