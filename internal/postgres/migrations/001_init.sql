CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    first_name TEXT        NOT NULL,
    last_name  TEXT        NOT NULL,
    phone      TEXT        NOT NULL UNIQUE,
    email      TEXT        UNIQUE,
    role       TEXT        NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- At most one active code per phone; requesting a new one replaces it.
CREATE TABLE otp_codes (
    phone        TEXT        PRIMARY KEY,
    code_hash    TEXT        NOT NULL,
    attempts     INTEGER     NOT NULL DEFAULT 0,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_sent_at TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE addresses (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    receiver_name  TEXT        NOT NULL,
    receiver_phone TEXT        NOT NULL,
    province       TEXT        NOT NULL,
    city           TEXT        NOT NULL,
    address_line   TEXT        NOT NULL,
    postal_code    TEXT        NOT NULL,
    is_default     BOOLEAN     NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_addresses_user_id ON addresses (user_id);

CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    price       BIGINT      NOT NULL CHECK (price >= 0),
    stock       INTEGER     NOT NULL CHECK (stock >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Order status only ever moves awaiting_payment -> paid, or
-- awaiting_payment -> cancelled (via the payment gateway confirm step).
CREATE TABLE orders (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 BIGINT      NOT NULL REFERENCES users (id),
    status                  TEXT        NOT NULL DEFAULT 'awaiting_payment'
                                CHECK (status IN ('awaiting_payment', 'paid', 'cancelled')),
    total_amount            BIGINT      NOT NULL CHECK (total_amount >= 0),
    -- Snapshot of the address chosen at checkout, so later address edits or
    -- deletion never change what a past order says it was shipped to.
    shipping_receiver_name  TEXT        NOT NULL,
    shipping_receiver_phone TEXT        NOT NULL,
    shipping_province       TEXT        NOT NULL,
    shipping_city           TEXT        NOT NULL,
    shipping_address_line   TEXT        NOT NULL,
    shipping_postal_code    TEXT        NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders (user_id, id DESC);

CREATE TABLE order_items (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT  NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id   BIGINT  NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    product_name TEXT    NOT NULL,
    quantity     INTEGER NOT NULL CHECK (quantity > 0),
    unit_price   BIGINT  NOT NULL CHECK (unit_price >= 0)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);

-- One payment session per order. A cancelled/expired session is terminal:
-- the shopper starts a fresh order rather than retrying the same one.
CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT      NOT NULL UNIQUE REFERENCES orders (id) ON DELETE CASCADE,
    authority  TEXT        NOT NULL UNIQUE,
    amount     BIGINT      NOT NULL CHECK (amount >= 0),
    status     TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled')),
    ref_id     TEXT,
    paid_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_authority ON payments (authority);
