package service

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

type SiteContentInput struct {
	ContactPhone, ContactEmail, ContactAddress, ContactHours, InstagramURL string
	AboutContent, TermsContent                                             string
}

type SiteContentService struct{ repository domain.SiteContentRepository }

func NewSiteContentService(repository domain.SiteContentRepository) *SiteContentService {
	return &SiteContentService{repository: repository}
}

func (s *SiteContentService) Get(ctx context.Context) (*domain.SiteContent, error) {
	content, err := s.repository.GetSiteContent(ctx)
	if err != nil {
		return nil, err
	}
	content.AboutContent = richtext.Sanitize(content.AboutContent)
	content.TermsContent = richtext.Sanitize(content.TermsContent)
	return content, nil
}

func (s *SiteContentService) Update(ctx context.Context, input SiteContentInput) (*domain.SiteContent, error) {
	content := &domain.SiteContent{
		ContactPhone: strings.TrimSpace(input.ContactPhone), ContactEmail: strings.TrimSpace(input.ContactEmail),
		ContactAddress: strings.TrimSpace(input.ContactAddress), ContactHours: strings.TrimSpace(input.ContactHours),
		InstagramURL: strings.TrimSpace(input.InstagramURL), AboutContent: richtext.Sanitize(input.AboutContent),
		TermsContent: richtext.Sanitize(input.TermsContent),
	}
	if len([]rune(content.ContactPhone)) > 40 || len([]rune(content.ContactEmail)) > 254 ||
		len([]rune(content.ContactAddress)) > 500 || len([]rune(content.ContactHours)) > 120 {
		return nil, fmt.Errorf("%w: contact details are too long", domain.ErrInvalidInput)
	}
	if content.ContactEmail != "" {
		address, err := mail.ParseAddress(content.ContactEmail)
		if err != nil || address.Address != content.ContactEmail {
			return nil, fmt.Errorf("%w: contact_email is invalid", domain.ErrInvalidInput)
		}
	}
	if content.InstagramURL != "" {
		parsed, err := url.ParseRequestURI(content.InstagramURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return nil, fmt.Errorf("%w: instagram_url must be an http or https URL", domain.ErrInvalidInput)
		}
	}
	if len([]rune(richtext.PlainText(content.AboutContent))) > 50000 || len(content.AboutContent) > 200000 ||
		len([]rune(richtext.PlainText(content.TermsContent))) > 50000 || len(content.TermsContent) > 200000 {
		return nil, fmt.Errorf("%w: page content must contain at most 50000 text characters", domain.ErrInvalidInput)
	}
	if err := s.repository.UpdateSiteContent(ctx, content); err != nil {
		return nil, err
	}
	return s.Get(ctx)
}
