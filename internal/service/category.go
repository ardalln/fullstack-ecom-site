package service

import (
	"context"
	"fmt"
	"strings"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

type CategoryInput struct {
	Name           string
	Slug           string
	Description    string
	SEOTitle       string
	SEODescription string
	HomeTitle      string
	HomeImageURL   string
	ShowOnHome     bool
	ParentID       *int64
}

type CategoryService struct{ categories domain.CategoryRepository }

func NewCategoryService(categories domain.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return s.categories.GetBySlug(ctx, slug)
}

func (s *CategoryService) Create(ctx context.Context, input CategoryInput) (*domain.Category, error) {
	category, err := normalizeCategory(input)
	if err != nil {
		return nil, err
	}
	if err := s.categories.ValidateParent(ctx, 0, category.ParentID); err != nil {
		return nil, err
	}
	if err := s.categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return s.categories.GetByID(ctx, category.ID)
}

func (s *CategoryService) Update(ctx context.Context, id int64, input CategoryInput) (*domain.Category, error) {
	category, err := normalizeCategory(input)
	if err != nil {
		return nil, err
	}
	category.ID = id
	if err := s.categories.ValidateParent(ctx, id, category.ParentID); err != nil {
		return nil, err
	}
	if err := s.categories.Update(ctx, category); err != nil {
		return nil, err
	}
	return s.categories.GetByID(ctx, category.ID)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	return s.categories.Delete(ctx, id)
}

func normalizeCategory(input CategoryInput) (*domain.Category, error) {
	name := strings.TrimSpace(input.Name)
	description := richtext.Sanitize(input.Description)
	seoTitle, seoDescription := strings.TrimSpace(input.SEOTitle), strings.TrimSpace(input.SEODescription)
	homeTitle, homeImageURL := strings.TrimSpace(input.HomeTitle), strings.TrimSpace(input.HomeImageURL)
	if name == "" {
		return nil, fmt.Errorf("%w: category name is required", domain.ErrInvalidInput)
	}
	if len([]rune(richtext.PlainText(description))) > 20000 || len([]rune(seoTitle)) > 200 || len([]rune(seoDescription)) > 320 || len([]rune(homeTitle)) > 120 {
		return nil, fmt.Errorf("%w: category description or SEO fields are too long", domain.ErrInvalidInput)
	}
	if homeImageURL != "" {
		if err := validateImageURL(homeImageURL); err != nil {
			return nil, err
		}
	}
	slug := makeEntitySlug(input.Slug, name)
	if slug == "" {
		return nil, fmt.Errorf("%w: category slug must contain letters or numbers", domain.ErrInvalidInput)
	}
	return &domain.Category{Name: name, Slug: slug, Description: description, SEOTitle: seoTitle, SEODescription: seoDescription, HomeTitle: homeTitle, HomeImageURL: homeImageURL, ShowOnHome: input.ShowOnHome, ParentID: input.ParentID}, nil
}
