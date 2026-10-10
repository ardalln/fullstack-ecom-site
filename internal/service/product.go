package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

// ProductInput is the data needed to create or replace a product.
type ProductInput struct {
	Name          string
	Slug          string
	Description   string
	ImageURL      string
	ImageURLs     []string
	Attributes    map[string]string
	VariantLabel  string
	Variants      []domain.ProductVariant
	CategoryID    *int64
	BrandID       int64
	Price         int64
	DiscountPrice int64
	Stock         int
	IsActive      *bool
	IsPopular     bool
}

func (in ProductInput) validate() error {
	switch {
	case strings.TrimSpace(in.Name) == "":
		return fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	case len(in.Variants) == 0 && in.Price <= 0:
		return fmt.Errorf("%w: price must be greater than zero", domain.ErrInvalidInput)
	case in.Stock < 0:
		return fmt.Errorf("%w: stock must not be negative", domain.ErrInvalidInput)
	case len([]rune(richtext.PlainText(in.Description))) > 20000:
		return fmt.Errorf("%w: product description must be at most 20000 characters", domain.ErrInvalidInput)
	case in.DiscountPrice < 0:
		return fmt.Errorf("%w: discount price must not be negative", domain.ErrInvalidInput)
	case in.CategoryID == nil || *in.CategoryID <= 0:
		return fmt.Errorf("%w: category_id is required", domain.ErrInvalidInput)
	case in.BrandID <= 0:
		return fmt.Errorf("%w: brand_id is required", domain.ErrInvalidInput)
	}
	if len(in.ImageURLs) > 8 {
		return fmt.Errorf("%w: at most 8 product images are allowed", domain.ErrInvalidInput)
	}
	if len(in.Variants) > 20 {
		return fmt.Errorf("%w: at most 20 product variants are allowed", domain.ErrInvalidInput)
	}
	if len(in.Variants) > 0 && strings.TrimSpace(in.VariantLabel) == "" {
		return fmt.Errorf("%w: variant_label is required when product variants are configured", domain.ErrInvalidInput)
	}
	if len(in.Variants) > 0 && in.DiscountPrice > 0 {
		return fmt.Errorf("%w: set discounts on individual variants instead of the product", domain.ErrInvalidInput)
	}
	if len(in.Variants) == 0 && in.DiscountPrice > 0 && in.DiscountPrice >= in.Price {
		return fmt.Errorf("%w: discount price must be lower than regular price", domain.ErrInvalidInput)
	}
	if len([]rune(strings.TrimSpace(in.VariantLabel))) > 50 {
		return fmt.Errorf("%w: variant_label must be at most 50 characters", domain.ErrInvalidInput)
	}
	seenVariants := make(map[string]struct{}, len(in.Variants))
	seenVariantIDs := make(map[string]struct{}, len(in.Variants))
	maxStock := int(^uint(0) >> 1)
	stockTotal := 0
	for _, variant := range in.Variants {
		name := strings.TrimSpace(variant.Name)
		if name == "" || len([]rune(name)) > 80 {
			return fmt.Errorf("%w: variant names must be 1-80 characters", domain.ErrInvalidInput)
		}
		if _, exists := seenVariants[strings.ToLower(name)]; exists {
			return fmt.Errorf("%w: variant names must be unique", domain.ErrInvalidInput)
		}
		seenVariants[strings.ToLower(name)] = struct{}{}
		if variant.Price <= 0 || variant.Stock < 0 || variant.Stock > maxStock-stockTotal {
			return fmt.Errorf("%w: each variant needs a positive price and non-negative stock", domain.ErrInvalidInput)
		}
		if variant.DiscountPrice < 0 || (variant.DiscountPrice > 0 && variant.DiscountPrice >= variant.Price) {
			return fmt.Errorf("%w: variant discount price must be lower than its regular price", domain.ErrInvalidInput)
		}
		stockTotal += variant.Stock
		if variant.ID != "" {
			if len(variant.ID) > 80 {
				return fmt.Errorf("%w: variant id is invalid", domain.ErrInvalidInput)
			}
			if _, exists := seenVariantIDs[variant.ID]; exists {
				return fmt.Errorf("%w: variant ids must be unique", domain.ErrInvalidInput)
			}
			seenVariantIDs[variant.ID] = struct{}{}
		}
		if strings.TrimSpace(variant.ImageURL) != "" {
			if err := validateImageURL(variant.ImageURL); err != nil {
				return err
			}
		}
	}
	images := append([]string(nil), in.ImageURLs...)
	if len(images) == 0 && strings.TrimSpace(in.ImageURL) != "" {
		images = append(images, strings.TrimSpace(in.ImageURL))
	}
	for _, image := range images {
		if err := validateImageURL(image); err != nil {
			return err
		}
	}
	if len(in.Attributes) > 20 {
		return fmt.Errorf("%w: at most 20 product attributes are allowed", domain.ErrInvalidInput)
	}
	for key, value := range in.Attributes {
		if strings.TrimSpace(key) == "" || len([]rune(key)) > 50 || len([]rune(value)) > 250 {
			return fmt.Errorf("%w: product attribute names must be 1-50 characters and values at most 250", domain.ErrInvalidInput)
		}
	}
	return nil
}

func validateImageURL(raw string) error {
	imageURL := strings.TrimSpace(raw)
	if strings.HasPrefix(imageURL, "/uploads/") {
		filename := strings.TrimPrefix(imageURL, "/uploads/")
		if filename != "" && !strings.ContainsAny(filename, "/\\") && !strings.Contains(filename, "..") &&
			(strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".webp")) {
			return nil
		}
	}
	parsed, err := url.Parse(imageURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return fmt.Errorf("%w: images must be uploaded or use an absolute HTTP/HTTPS URL", domain.ErrInvalidInput)
	}
	return nil
}

type ProductService struct {
	products   domain.ProductRepository
	categories domain.CategoryRepository
	brands     domain.BrandRepository
}

func NewProductService(products domain.ProductRepository, categories domain.CategoryRepository, brands domain.BrandRepository) *ProductService {
	return &ProductService{products: products, categories: categories, brands: brands}
}

func (s *ProductService) Create(ctx context.Context, in ProductInput) (*domain.Product, error) {
	in.Description = richtext.Sanitize(in.Description)
	if err := in.validate(); err != nil {
		return nil, err
	}
	if err := s.validateCategory(ctx, in.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateBrand(ctx, in.BrandID); err != nil {
		return nil, err
	}
	var err error
	in, err = normalizeProductInput(in)
	if err != nil {
		return nil, err
	}
	p := &domain.Product{
		Name:          strings.TrimSpace(in.Name),
		Slug:          makeSlug(in.Slug, in.Name),
		Description:   strings.TrimSpace(in.Description),
		ImageURL:      strings.TrimSpace(in.ImageURL),
		ImageURLs:     in.ImageURLs,
		Attributes:    in.Attributes,
		VariantLabel:  in.VariantLabel,
		Variants:      in.Variants,
		CategoryID:    in.CategoryID,
		BrandID:       in.BrandID,
		Price:         in.Price,
		DiscountPrice: in.DiscountPrice,
		Stock:         in.Stock,
		IsActive:      activeValue(in.IsActive, true),
		IsPopular:     in.IsPopular,
	}
	if err := s.products.Create(ctx, p); err != nil {
		return nil, err
	}
	return s.products.GetByID(ctx, p.ID)
}

func (s *ProductService) Get(ctx context.Context, id int64) (*domain.Product, error) {
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !product.IsActive {
		return nil, domain.ErrNotFound
	}
	return product, nil
}

func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := s.products.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !product.IsActive {
		return nil, domain.ErrNotFound
	}
	return product, nil
}

func (s *ProductService) List(ctx context.Context, search string, categoryID, brandID *int64, page, limit int) ([]domain.Product, int, error) {
	if categoryID != nil {
		if _, err := s.categories.GetByID(ctx, *categoryID); err != nil {
			return nil, 0, err
		}
	}
	if brandID != nil {
		if _, err := s.brands.GetByID(ctx, *brandID); err != nil {
			return nil, 0, err
		}
	}
	return s.products.List(ctx, search, categoryID, brandID, limit, (page-1)*limit)
}

func (s *ProductService) ListPopular(ctx context.Context, page, limit int) ([]domain.Product, int, error) {
	return s.products.ListPopular(ctx, limit, (page-1)*limit)
}

func (s *ProductService) ListAdmin(ctx context.Context, search string, categoryID, brandID *int64, page, limit int) ([]domain.Product, int, error) {
	if categoryID != nil {
		if _, err := s.categories.GetByID(ctx, *categoryID); err != nil {
			return nil, 0, err
		}
	}
	if brandID != nil {
		if _, err := s.brands.GetByID(ctx, *brandID); err != nil {
			return nil, 0, err
		}
	}
	return s.products.ListAdmin(ctx, search, categoryID, brandID, limit, (page-1)*limit)
}

func (s *ProductService) Update(ctx context.Context, id int64, in ProductInput) (*domain.Product, error) {
	in.Description = richtext.Sanitize(in.Description)
	if err := in.validate(); err != nil {
		return nil, err
	}
	if err := s.validateCategory(ctx, in.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateBrand(ctx, in.BrandID); err != nil {
		return nil, err
	}
	if in.IsActive == nil {
		current, err := s.products.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		in.IsActive = &current.IsActive
	}
	var err error
	in, err = normalizeProductInput(in)
	if err != nil {
		return nil, err
	}
	p := &domain.Product{
		ID:            id,
		Name:          strings.TrimSpace(in.Name),
		Slug:          makeSlug(in.Slug, in.Name),
		Description:   strings.TrimSpace(in.Description),
		ImageURL:      strings.TrimSpace(in.ImageURL),
		ImageURLs:     in.ImageURLs,
		Attributes:    in.Attributes,
		VariantLabel:  in.VariantLabel,
		Variants:      in.Variants,
		CategoryID:    in.CategoryID,
		BrandID:       in.BrandID,
		Price:         in.Price,
		DiscountPrice: in.DiscountPrice,
		Stock:         in.Stock,
		IsActive:      activeValue(in.IsActive, true),
		IsPopular:     in.IsPopular,
	}
	if err := s.products.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.products.GetByID(ctx, p.ID)
}

func (s *ProductService) SetActive(ctx context.Context, id int64, active bool) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.products.SetActive(ctx, id, active)
}

func activeValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func (s *ProductService) validateBrand(ctx context.Context, id int64) error {
	if _, err := s.brands.GetByID(ctx, id); err != nil {
		return err
	}
	return nil
}

func normalizeProductInput(in ProductInput) (ProductInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Slug = makeSlug(in.Slug, in.Name)
	in.Description = richtext.Sanitize(in.Description)
	in.ImageURLs = append([]string(nil), in.ImageURLs...)
	in.Variants = append([]domain.ProductVariant(nil), in.Variants...)
	for i := range in.Variants {
		if in.Variants[i].ID == "" {
			id, err := makeVariantID()
			if err != nil {
				return ProductInput{}, err
			}
			in.Variants[i].ID = id
		}
		in.Variants[i].Name = strings.TrimSpace(in.Variants[i].Name)
		in.Variants[i].ImageURL = strings.TrimSpace(in.Variants[i].ImageURL)
	}
	in.VariantLabel = strings.TrimSpace(in.VariantLabel)
	if len(in.Variants) > 0 {
		in.Price = in.Variants[0].Price
		in.Stock = 0
		for _, variant := range in.Variants {
			if variant.Price < in.Price {
				in.Price = variant.Price
			}
			in.Stock += variant.Stock
		}
	} else {
		in.VariantLabel = ""
	}
	if len(in.ImageURLs) == 0 && strings.TrimSpace(in.ImageURL) != "" {
		in.ImageURLs = []string{strings.TrimSpace(in.ImageURL)}
	}
	if in.ImageURLs == nil {
		in.ImageURLs = []string{}
	}
	if len(in.ImageURLs) > 0 {
		in.ImageURL = in.ImageURLs[0]
	}
	if in.Attributes == nil {
		in.Attributes = map[string]string{}
	}
	return in, nil
}

func makeVariantID() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate product variant id: %w", err)
	}
	return hex.EncodeToString(token[:]), nil
}

func makeSlug(raw, name string) string {
	if strings.TrimSpace(raw) == "" {
		raw = name
	}
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if replacement, ok := persianSlug[r]; ok {
			for _, rr := range replacement {
				b.WriteRune(rr)
			}
			separator = false
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			separator = false
			continue
		}
		if r == '-' || r == '_' || unicode.IsSpace(r) {
			if b.Len() > 0 && !separator {
				b.WriteByte('-')
				separator = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug != "" {
		return slug
	}
	var token [6]byte
	if _, err := rand.Read(token[:]); err == nil {
		return "product-" + hex.EncodeToString(token[:])
	}
	return "product-item"
}

var persianSlug = map[rune]string{
	'ا': "a", 'آ': "a", 'ب': "b", 'پ': "p", 'ت': "t", 'ث': "s", 'ج': "j", 'چ': "ch", 'ح': "h", 'خ': "kh", 'د': "d", 'ذ': "z", 'ر': "r", 'ز': "z", 'ژ': "zh", 'س': "s", 'ش': "sh", 'ص': "s", 'ض': "z", 'ط': "t", 'ظ': "z", 'ع': "a", 'غ': "gh", 'ف': "f", 'ق': "gh", 'ک': "k", 'ك': "k", 'گ': "g", 'ل': "l", 'م': "m", 'ن': "n", 'و': "v", 'ؤ': "v", 'ه': "h", 'ۀ': "h", 'ة': "h", 'ی': "y", 'ي': "y", 'ئ': "y", 'ى': "y",
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.products.Delete(ctx, id)
}

func (s *ProductService) validateCategory(ctx context.Context, categoryID *int64) error {
	if categoryID == nil {
		return nil
	}
	_, err := s.categories.GetByID(ctx, *categoryID)
	return err
}
