ALTER TABLE products ADD COLUMN variant_label TEXT NOT NULL DEFAULT '';
ALTER TABLE order_items ADD COLUMN variant_id TEXT NOT NULL DEFAULT '';
ALTER TABLE order_items ADD COLUMN variant_label TEXT NOT NULL DEFAULT '';

-- Older options shared the product's price and stock. Preserve the product's
-- total quantity while assigning the existing stock across those options.
WITH expanded AS (
    SELECT p.id, p.price, p.stock, p.variants,
           option.ordinality, option.item,
           jsonb_array_length(p.variants) AS option_count
    FROM products p
    CROSS JOIN LATERAL jsonb_array_elements(p.variants) WITH ORDINALITY AS option(item, ordinality)
    WHERE jsonb_array_length(p.variants) > 0
), rebuilt AS (
    SELECT id,
           jsonb_agg(
               item || jsonb_build_object(
                   'id', COALESCE(NULLIF(item->>'id', ''), 'legacy-' || id::text || '-' || ordinality::text),
                   'price', COALESCE(NULLIF(item->>'price', '')::BIGINT, price),
                   'stock', COALESCE(
                       NULLIF(item->>'stock', '')::INTEGER,
                       stock / option_count + CASE WHEN ordinality <= stock % option_count THEN 1 ELSE 0 END
                   )
               ) ORDER BY ordinality
           ) AS variants
    FROM expanded
    GROUP BY id
)
UPDATE products p
SET variants = rebuilt.variants,
    variant_label = CASE WHEN p.variant_label = '' THEN 'مدل' ELSE p.variant_label END
FROM rebuilt
WHERE rebuilt.id = p.id;

-- Existing pending orders already stored the selected option's name.
UPDATE order_items oi
SET variant_id = option.item->>'id',
    variant_label = CASE WHEN p.variant_label = '' THEN 'مدل' ELSE p.variant_label END
FROM products p
CROSS JOIN LATERAL jsonb_array_elements(p.variants) AS option(item)
WHERE oi.product_id = p.id
  AND oi.variant_name <> ''
  AND lower(option.item->>'name') = lower(oi.variant_name);
