package service

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"shop-api/internal/domain"
)

type CommerceService struct {
	banners  domain.BannerRepository
	shipping domain.ShippingMethodRepository
	coupons  domain.CouponRepository
}

const maxTomanAmount int64 = 1_000_000_000_000_000

func NewCommerceService(b domain.BannerRepository, s domain.ShippingMethodRepository, c domain.CouponRepository) *CommerceService {
	return &CommerceService{banners: b, shipping: s, coupons: c}
}

func (s *CommerceService) Banners(ctx context.Context, admin bool) ([]domain.Banner, error) {
	if admin {
		return s.banners.ListAll(ctx)
	}
	return s.banners.ListActive(ctx)
}
func (s *CommerceService) SaveBanner(ctx context.Context, b *domain.Banner) (*domain.Banner, error) {
	b.Title = strings.TrimSpace(b.Title)
	b.Subtitle = strings.TrimSpace(b.Subtitle)
	b.DesktopImageURL = strings.TrimSpace(b.DesktopImageURL)
	b.MobileImageURL = strings.TrimSpace(b.MobileImageURL)
	b.LinkURL = strings.TrimSpace(b.LinkURL)
	b.ButtonLabel = strings.TrimSpace(b.ButtonLabel)
	if !safeBannerURL(b.DesktopImageURL, false) || !safeBannerURL(b.MobileImageURL, false) || !safeBannerURL(b.LinkURL, true) || len([]rune(b.Title)) > 160 || len([]rune(b.Subtitle)) > 300 || len([]rune(b.ButtonLabel)) > 50 || b.SortOrder < 0 {
		return nil, fmt.Errorf("%w: banner image URLs and valid text are required", domain.ErrInvalidInput)
	}
	var err error
	if b.ID == 0 {
		err = s.banners.Create(ctx, b)
	} else {
		err = s.banners.Update(ctx, b)
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// safeBannerURL only permits same-origin paths or HTTPS URLs for banner
// resources and destinations. In particular, javascript: and protocol-relative
// links are rejected to prevent stored active-content URLs.
func safeBannerURL(raw string, allowEmpty bool) bool {
	if raw == "" {
		return allowEmpty
	}
	if strings.ContainsAny(raw, "\\\r\n\x00") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil {
		return false
	}
	if u.Scheme == "" {
		return (strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//")) || (strings.HasPrefix(raw, "#") && allowEmpty)
	}
	return strings.EqualFold(u.Scheme, "https") && u.Host != ""
}
func (s *CommerceService) DeleteBanner(ctx context.Context, id int64) error {
	return s.banners.Delete(ctx, id)
}
func (s *CommerceService) ShippingMethods(ctx context.Context, province, city string, admin bool) ([]domain.ShippingMethod, error) {
	if admin {
		return s.shipping.ListAll(ctx)
	}
	if strings.TrimSpace(province) == "" || strings.TrimSpace(city) == "" {
		return nil, fmt.Errorf("%w: province and city are required", domain.ErrInvalidInput)
	}
	return s.shipping.ListAvailable(ctx, province, city)
}
func (s *CommerceService) SaveShipping(ctx context.Context, m *domain.ShippingMethod) (*domain.ShippingMethod, error) {
	m.Name = strings.TrimSpace(m.Name)
	m.Description = strings.TrimSpace(m.Description)
	m.Province = strings.TrimSpace(m.Province)
	m.City = strings.TrimSpace(m.City)
	if m.Name == "" || len([]rune(m.Name)) > 100 || m.Price < 0 || m.Price > maxTomanAmount || m.SortOrder < 0 || (m.City != "" && m.Province == "") {
		return nil, fmt.Errorf("%w: shipping method fields are invalid", domain.ErrInvalidInput)
	}
	var err error
	if m.ID == 0 {
		err = s.shipping.Create(ctx, m)
	} else {
		err = s.shipping.Update(ctx, m)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}
func (s *CommerceService) DeleteShipping(ctx context.Context, id int64) error {
	return s.shipping.Delete(ctx, id)
}

func (s *CommerceService) Coupons(ctx context.Context) ([]domain.Coupon, error) {
	return s.coupons.ListAll(ctx)
}
func (s *CommerceService) SaveCoupon(ctx context.Context, c *domain.Coupon) (*domain.Coupon, error) {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	c.Kind = strings.ToLower(strings.TrimSpace(c.Kind))
	if c.Code == "" || len(c.Code) > 50 || c.Value <= 0 || c.Value > maxTomanAmount || c.MinimumSubtotal < 0 || c.MinimumSubtotal > maxTomanAmount || (c.Kind != "percent" && c.Kind != "fixed") || (c.Kind == "percent" && c.Value > 100) || (c.MaximumDiscount != nil && (*c.MaximumDiscount <= 0 || *c.MaximumDiscount > maxTomanAmount)) || (c.UsageLimit != nil && *c.UsageLimit <= 0) || (c.StartsAt != nil && c.EndsAt != nil && !c.EndsAt.After(*c.StartsAt)) {
		return nil, fmt.Errorf("%w: coupon fields are invalid", domain.ErrInvalidInput)
	}
	var err error
	if c.ID == 0 {
		err = s.coupons.Create(ctx, c)
	} else {
		err = s.coupons.Update(ctx, c)
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (s *CommerceService) DeleteCoupon(ctx context.Context, id int64) error {
	return s.coupons.Delete(ctx, id)
}

func (s *CommerceService) PreviewCoupon(ctx context.Context, code string, subtotal int64) (*domain.Coupon, int64, error) {
	if subtotal < 0 || subtotal > maxTomanAmount {
		return nil, 0, domain.ErrInvalidInput
	}
	c, err := s.coupons.GetByCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, 0, domain.ErrCouponInvalid
	}
	now := time.Now()
	if !c.IsActive || (c.StartsAt != nil && now.Before(*c.StartsAt)) || (c.EndsAt != nil && !now.Before(*c.EndsAt)) || (c.UsageLimit != nil && c.UsedCount >= *c.UsageLimit) {
		return nil, 0, domain.ErrCouponInvalid
	}
	discount, err := couponDiscount(c, subtotal)
	if err != nil {
		return nil, 0, err
	}
	return c, discount, nil
}
func couponDiscount(c *domain.Coupon, subtotal int64) (int64, error) {
	if subtotal < c.MinimumSubtotal {
		return 0, domain.ErrCouponMinimum
	}
	var discount int64
	if c.Kind == "percent" {
		n := new(big.Int).Mul(big.NewInt(subtotal), big.NewInt(c.Value))
		n.Div(n, big.NewInt(100))
		if !n.IsInt64() {
			return 0, domain.ErrInvalidInput
		}
		discount = n.Int64()
	} else {
		discount = c.Value
	}
	if c.MaximumDiscount != nil && discount > *c.MaximumDiscount {
		discount = *c.MaximumDiscount
	}
	if discount > subtotal {
		discount = subtotal
	}
	return discount, nil
}
