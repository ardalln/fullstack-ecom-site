package service

import (
	"context"
	"fmt"
	"strings"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

type BrandInput struct {
	Name, Slug, Description  string
	SEOTitle, SEODescription string
}

type BrandService struct{ brands domain.BrandRepository }

func NewBrandService(brands domain.BrandRepository) *BrandService {
	return &BrandService{brands: brands}
}
func (s *BrandService) List(ctx context.Context) ([]domain.Brand, error) { return s.brands.List(ctx) }
func (s *BrandService) GetBySlug(ctx context.Context, slug string) (*domain.Brand, error) {
	return s.brands.GetBySlug(ctx, slug)
}
func (s *BrandService) Create(ctx context.Context, input BrandInput) (*domain.Brand, error) {
	brand, err := normalizeBrand(input)
	if err != nil {
		return nil, err
	}
	if err := s.brands.Create(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}
func (s *BrandService) Update(ctx context.Context, id int64, input BrandInput) (*domain.Brand, error) {
	brand, err := normalizeBrand(input)
	if err != nil {
		return nil, err
	}
	brand.ID = id
	if err := s.brands.Update(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}
func (s *BrandService) Delete(ctx context.Context, id int64) error { return s.brands.Delete(ctx, id) }

func normalizeBrand(input BrandInput) (*domain.Brand, error) {
	name, description := strings.TrimSpace(input.Name), richtext.Sanitize(input.Description)
	seoTitle, seoDescription := strings.TrimSpace(input.SEOTitle), strings.TrimSpace(input.SEODescription)
	if name == "" {
		return nil, fmt.Errorf("%w: brand name is required", domain.ErrInvalidInput)
	}
	if len([]rune(richtext.PlainText(description))) > 20000 || len([]rune(seoTitle)) > 200 || len([]rune(seoDescription)) > 320 {
		return nil, fmt.Errorf("%w: brand description or SEO fields are too long", domain.ErrInvalidInput)
	}
	slug := makeEntitySlug(input.Slug, name)
	if slug == "" {
		return nil, fmt.Errorf("%w: brand slug must contain letters or numbers", domain.ErrInvalidInput)
	}
	return &domain.Brand{Name: name, Slug: slug, Description: description, SEOTitle: seoTitle, SEODescription: seoDescription}, nil
}
