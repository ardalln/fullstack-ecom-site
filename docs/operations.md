# Operations

Unpaid orders reserve stock and coupon usage for `ORDER_PAYMENT_TTL` (30 minutes by default). A background task checks for expired orders once a minute. It locks each order and, in one database transaction, cancels any pending payment, cancels the order, restores product or variant stock, and releases the coupon reservation. The expiry timestamp is set when the order is created, so a payment attempt cannot extend the stock reservation. Existing orders are assigned an expiry of 30 minutes after their creation when migration 012 is applied.

OTP and analytics request limits use PostgreSQL-backed token buckets, so all application replicas share the same limits. Their state is pruned periodically. Configure `TRUSTED_PROXIES` with comma-separated IP addresses or CIDRs only when the listed proxies overwrite `X-Forwarded-For`; leave it empty when clients connect directly. The application trusts forwarded client IPs only through those configured proxy ranges.

Frontend dependencies are restored from `frontend/package-lock.json`; `node_modules/` and ZIP archives are intentionally excluded from Git.
