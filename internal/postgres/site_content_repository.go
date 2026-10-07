package postgres

import (
	"context"
	"fmt"

	"shop-api/internal/domain"
)

type SiteContentRepository struct{ db DBTX }

func NewSiteContentRepository(db DBTX) *SiteContentRepository {
	return &SiteContentRepository{db: db}
}

func (r *SiteContentRepository) GetSiteContent(ctx context.Context) (*domain.SiteContent, error) {
	var content domain.SiteContent
	err := r.db.QueryRow(ctx, `SELECT contact_phone, contact_email, contact_address, contact_hours,
		instagram_url, about_content, terms_content, updated_at FROM site_content WHERE id=1`).Scan(
		&content.ContactPhone, &content.ContactEmail, &content.ContactAddress, &content.ContactHours,
		&content.InstagramURL, &content.AboutContent, &content.TermsContent, &content.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get site content: %w", err)
	}
	return &content, nil
}

func (r *SiteContentRepository) UpdateSiteContent(ctx context.Context, content *domain.SiteContent) error {
	err := r.db.QueryRow(ctx, `UPDATE site_content SET contact_phone=$1, contact_email=$2,
		contact_address=$3, contact_hours=$4, instagram_url=$5, about_content=$6,
		terms_content=$7, updated_at=now() WHERE id=1 RETURNING updated_at`,
		content.ContactPhone, content.ContactEmail, content.ContactAddress, content.ContactHours,
		content.InstagramURL, content.AboutContent, content.TermsContent).Scan(&content.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update site content: %w", err)
	}
	return nil
}
