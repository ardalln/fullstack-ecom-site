ALTER TABLE orders ADD COLUMN payment_expires_at TIMESTAMPTZ;
UPDATE orders
SET payment_expires_at = created_at + INTERVAL '30 minutes'
WHERE payment_expires_at IS NULL;
ALTER TABLE orders ALTER COLUMN payment_expires_at SET NOT NULL;

CREATE INDEX idx_orders_expiring_payment
    ON orders (payment_expires_at, id)
    WHERE status = 'awaiting_payment';

CREATE TABLE request_rate_limits (
    client_key CHAR(64) PRIMARY KEY,
    tokens_milli BIGINT NOT NULL CHECK (tokens_milli >= 0),
    last_allowed BOOLEAN NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_request_rate_limits_updated_at ON request_rate_limits (updated_at);
