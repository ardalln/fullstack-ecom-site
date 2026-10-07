package postgres

import (
	"context"
	"fmt"

	"shop-api/internal/domain"
)

type AnalyticsRepository struct{ db DBTX }

func NewAnalyticsRepository(db DBTX) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) Heartbeat(ctx context.Context, visitorID, path string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO analytics_live_visitors(visitor_id,current_path)
		VALUES($1,$2) ON CONFLICT(visitor_id) DO UPDATE
		SET current_path=EXCLUDED.current_path,last_seen=now()`, visitorID, path)
	if err != nil {
		return fmt.Errorf("record analytics heartbeat: %w", err)
	}
	_, _ = r.db.Exec(ctx, `DELETE FROM analytics_live_visitors WHERE last_seen < now() - interval '1 day'`)
	return nil
}

func (r *AnalyticsRepository) RecordPageView(ctx context.Context, viewID, visitorID, path string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO analytics_page_views(view_id,visitor_id,path)
		VALUES($1,$2,$3) ON CONFLICT(view_id) DO NOTHING`, viewID, visitorID, path)
	if err != nil {
		return fmt.Errorf("record analytics page view: %w", err)
	}
	return nil
}

func (r *AnalyticsRepository) GetAnalyticsOverview(ctx context.Context) (domain.AnalyticsOverview, error) {
	var result domain.AnalyticsOverview
	err := r.db.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM analytics_live_visitors WHERE last_seen > now() - interval '90 seconds'),
		(SELECT count(DISTINCT visitor_id) FROM analytics_page_views WHERE viewed_at >= date_trunc('day', now() AT TIME ZONE 'Asia/Tehran') AT TIME ZONE 'Asia/Tehran'),
		(SELECT count(DISTINCT visitor_id) FROM analytics_page_views),
		(SELECT count(*) FROM analytics_page_views WHERE viewed_at >= date_trunc('day', now() AT TIME ZONE 'Asia/Tehran') AT TIME ZONE 'Asia/Tehran')`).Scan(
		&result.ActiveOnline, &result.VisitorsToday, &result.UniqueVisitors,
		&result.PageViewsToday)
	if err != nil {
		return result, fmt.Errorf("read analytics overview: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT visitor_id::text,current_path,last_seen
		FROM analytics_live_visitors WHERE last_seen > now() - interval '90 seconds'
		ORDER BY last_seen DESC LIMIT 200`)
	if err != nil {
		return result, fmt.Errorf("list active analytics visitors: %w", err)
	}
	defer rows.Close()
	result.ActiveVisitors = make([]domain.AnalyticsVisitor, 0)
	for rows.Next() {
		var visitor domain.AnalyticsVisitor
		if err := rows.Scan(&visitor.VisitorID, &visitor.Path, &visitor.LastSeen); err != nil {
			return result, fmt.Errorf("scan active analytics visitor: %w", err)
		}
		result.ActiveVisitors = append(result.ActiveVisitors, visitor)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate active analytics visitors: %w", err)
	}
	return result, nil
}
