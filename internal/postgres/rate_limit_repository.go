package postgres

import (
	"context"
	"fmt"
	"time"
)

type RateLimitRepository struct{ db DBTX }

func NewRateLimitRepository(db DBTX) *RateLimitRepository {
	return &RateLimitRepository{db: db}
}

func (r *RateLimitRepository) Allow(ctx context.Context, key string, capacity int, refillInterval time.Duration) (bool, error) {
	if capacity < 1 || refillInterval <= 0 {
		return false, fmt.Errorf("rate limit capacity and refill interval must be positive")
	}
	const q = `
		INSERT INTO request_rate_limits (client_key, tokens_milli, last_allowed, updated_at)
		VALUES ($1, $2::bigint * 1000 - 1000, true, now())
		ON CONFLICT (client_key) DO UPDATE SET
			tokens_milli = CASE WHEN
				LEAST($2::bigint * 1000,
					request_rate_limits.tokens_milli +
					FLOOR(GREATEST(0, EXTRACT(EPOCH FROM (now() - request_rate_limits.updated_at))) * 1000 / $3)::bigint
				) >= 1000
			THEN LEAST($2::bigint * 1000,
					request_rate_limits.tokens_milli +
					FLOOR(GREATEST(0, EXTRACT(EPOCH FROM (now() - request_rate_limits.updated_at))) * 1000 / $3)::bigint
				) - 1000
			ELSE LEAST($2::bigint * 1000,
					request_rate_limits.tokens_milli +
					FLOOR(GREATEST(0, EXTRACT(EPOCH FROM (now() - request_rate_limits.updated_at))) * 1000 / $3)::bigint
				)
			END,
			last_allowed = LEAST($2::bigint * 1000,
				request_rate_limits.tokens_milli +
				FLOOR(GREATEST(0, EXTRACT(EPOCH FROM (now() - request_rate_limits.updated_at))) * 1000 / $3)::bigint
			) >= 1000,
			updated_at = now()
		RETURNING last_allowed`
	var allowed bool
	if err := r.db.QueryRow(ctx, q, key, capacity, refillInterval.Seconds()).Scan(&allowed); err != nil {
		return false, fmt.Errorf("check shared request rate limit: %w", err)
	}
	return allowed, nil
}

func (r *RateLimitRepository) DeleteExpired(ctx context.Context, before time.Time, limit int) (int64, error) {
	const q = `
		WITH expired AS (
			SELECT client_key
			FROM request_rate_limits
			WHERE updated_at < $1
			ORDER BY updated_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM request_rate_limits AS limits
		USING expired
		WHERE limits.client_key = expired.client_key`
	tag, err := r.db.Exec(ctx, q, before, limit)
	if err != nil {
		return 0, fmt.Errorf("delete expired request rate limits: %w", err)
	}
	return tag.RowsAffected(), nil
}
