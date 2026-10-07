package service

import (
	"context"
	"fmt"
	"strings"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

type BlogInput struct {
	Title          string
	Slug           string
	Summary        string
	Content        string
	CoverImageURL  string
	SEOTitle       string
	SEODescription string
	IsPublished    bool
}

type BlogService struct{ posts domain.BlogRepository }

func NewBlogService(posts domain.BlogRepository) *BlogService { return &BlogService{posts: posts} }

func (s *BlogService) Create(ctx context.Context, in BlogInput) (*domain.BlogPost, error) {
	post, err := normalizeBlogInput(in)
	if err != nil {
		return nil, err
	}
	if err := s.posts.Create(ctx, post); err != nil {
		return nil, err
	}
	return s.posts.GetByID(ctx, post.ID)
}

func (s *BlogService) Update(ctx context.Context, id int64, in BlogInput) (*domain.BlogPost, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	post, err := normalizeBlogInput(in)
	if err != nil {
		return nil, err
	}
	post.ID = id
	if err := s.posts.Update(ctx, post); err != nil {
		return nil, err
	}
	return s.posts.GetByID(ctx, id)
}

func (s *BlogService) GetPublished(ctx context.Context, slug string) (*domain.BlogPost, error) {
	post, err := s.posts.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !post.IsPublished {
		return nil, domain.ErrNotFound
	}
	return post, nil
}

func (s *BlogService) GetAdmin(ctx context.Context, id int64) (*domain.BlogPost, error) {
	if id <= 0 {
		return nil, domain.ErrNotFound
	}
	return s.posts.GetByID(ctx, id)
}

func (s *BlogService) ListPublished(ctx context.Context, search string, page, limit int) ([]domain.BlogPost, int, error) {
	return s.posts.ListPublished(ctx, strings.TrimSpace(search), limit, (page-1)*limit)
}

func (s *BlogService) ListAdmin(ctx context.Context, search string, page, limit int) ([]domain.BlogPost, int, error) {
	return s.posts.ListAdmin(ctx, strings.TrimSpace(search), limit, (page-1)*limit)
}

func (s *BlogService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.posts.Delete(ctx, id)
}

func normalizeBlogInput(in BlogInput) (*domain.BlogPost, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Summary = strings.TrimSpace(in.Summary)
	in.Content = richtext.Sanitize(in.Content)
	in.Slug = makeSlug(in.Slug, in.Title)
	in.CoverImageURL = strings.TrimSpace(in.CoverImageURL)
	in.SEOTitle = strings.TrimSpace(in.SEOTitle)
	in.SEODescription = strings.TrimSpace(in.SEODescription)
	switch {
	case in.Title == "" || len([]rune(in.Title)) > 200:
		return nil, fmt.Errorf("%w: title must contain 1-200 characters", domain.ErrInvalidInput)
	case in.Slug == "" || len(in.Slug) > 180:
		return nil, fmt.Errorf("%w: slug must contain 1-180 characters", domain.ErrInvalidInput)
	case len([]rune(in.Summary)) > 500:
		return nil, fmt.Errorf("%w: summary must be at most 500 characters", domain.ErrInvalidInput)
	case richtext.PlainText(in.Content) == "" || len([]rune(richtext.PlainText(in.Content))) > 50000 || len(in.Content) > 200000:
		return nil, fmt.Errorf("%w: content must contain 1-50000 text characters", domain.ErrInvalidInput)
	case len([]rune(in.SEOTitle)) > 200:
		return nil, fmt.Errorf("%w: seo_title must be at most 200 characters", domain.ErrInvalidInput)
	case len([]rune(in.SEODescription)) > 320:
		return nil, fmt.Errorf("%w: seo_description must be at most 320 characters", domain.ErrInvalidInput)
	}
	if in.CoverImageURL != "" {
		if err := validateImageURL(in.CoverImageURL); err != nil {
			return nil, err
		}
	}
	if in.SEOTitle == "" {
		in.SEOTitle = in.Title
	}
	if in.SEODescription == "" {
		in.SEODescription = in.Summary
	}
	return &domain.BlogPost{
		Title: in.Title, Slug: in.Slug, Summary: in.Summary, Content: in.Content,
		CoverImageURL: in.CoverImageURL, SEOTitle: in.SEOTitle,
		SEODescription: in.SEODescription, IsPublished: in.IsPublished,
	}, nil
}
