ALTER TABLE brands
    ADD COLUMN seo_title TEXT NOT NULL DEFAULT '' CHECK (length(seo_title) <= 200),
    ADD COLUMN seo_description TEXT NOT NULL DEFAULT '' CHECK (length(seo_description) <= 320);

ALTER TABLE categories
    ADD COLUMN seo_title TEXT NOT NULL DEFAULT '' CHECK (length(seo_title) <= 200),
    ADD COLUMN seo_description TEXT NOT NULL DEFAULT '' CHECK (length(seo_description) <= 320),
    ADD COLUMN home_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN home_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN show_on_home BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE categories SET show_on_home = TRUE WHERE parent_id IS NULL;

ALTER TABLE blog_posts DROP CONSTRAINT blog_posts_content_check;
ALTER TABLE blog_posts ADD CONSTRAINT blog_posts_content_check CHECK (length(content) BETWEEN 1 AND 200000);

CREATE TABLE site_content (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    contact_phone TEXT NOT NULL DEFAULT '',
    contact_email TEXT NOT NULL DEFAULT '',
    contact_address TEXT NOT NULL DEFAULT '',
    contact_hours TEXT NOT NULL DEFAULT '',
    instagram_url TEXT NOT NULL DEFAULT '',
    about_content TEXT NOT NULL DEFAULT '',
    terms_content TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO site_content (id, about_content, terms_content) VALUES (
    1,
    '<p>اینجا داستان و مسیر OUTSIDE را بنویسید.</p>',
    '<p>قوانین و شرایط استفاده از فروشگاه را اینجا منتشر کنید.</p>'
) ON CONFLICT (id) DO NOTHING;

CREATE TABLE analytics_live_visitors (
    visitor_id UUID PRIMARY KEY,
    current_path TEXT NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_analytics_live_visitors_last_seen ON analytics_live_visitors (last_seen DESC);

CREATE TABLE analytics_page_views (
    view_id UUID PRIMARY KEY,
    visitor_id UUID NOT NULL,
    path TEXT NOT NULL,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_analytics_page_views_viewed_at ON analytics_page_views (viewed_at DESC);
CREATE INDEX idx_analytics_page_views_visitor ON analytics_page_views (visitor_id, viewed_at DESC);
