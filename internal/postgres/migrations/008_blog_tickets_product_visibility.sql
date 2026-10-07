ALTER TABLE products ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE blog_posts (
    id               BIGSERIAL PRIMARY KEY,
    title            TEXT NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
    slug             TEXT NOT NULL,
    summary          TEXT NOT NULL DEFAULT '' CHECK (length(summary) <= 500),
    content          TEXT NOT NULL CHECK (length(content) BETWEEN 1 AND 50000),
    cover_image_url  TEXT NOT NULL DEFAULT '',
    seo_title        TEXT NOT NULL DEFAULT '' CHECK (length(seo_title) <= 200),
    seo_description  TEXT NOT NULL DEFAULT '' CHECK (length(seo_description) <= 320),
    is_published     BOOLEAN NOT NULL DEFAULT FALSE,
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_blog_posts_slug ON blog_posts (slug);
CREATE INDEX idx_blog_posts_published ON blog_posts (published_at DESC, id DESC) WHERE is_published;

CREATE TABLE support_tickets (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    subject     TEXT NOT NULL CHECK (length(subject) BETWEEN 1 AND 200),
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'answered', 'closed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_tickets_user ON support_tickets (user_id, updated_at DESC, id DESC);
CREATE INDEX idx_support_tickets_admin ON support_tickets (status, updated_at DESC, id DESC);

CREATE TABLE support_ticket_messages (
    id          BIGSERIAL PRIMARY KEY,
    ticket_id   BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    sender_name TEXT NOT NULL,
    sender_role TEXT NOT NULL CHECK (sender_role IN ('user', 'admin')),
    body        TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 5000),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_ticket_messages_ticket ON support_ticket_messages (ticket_id, id ASC);
