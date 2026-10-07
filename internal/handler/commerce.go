package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type CommerceHandler struct {
	svc    *service.CommerceService
	orders domain.OrderRepository
}

func NewCommerceHandler(s *service.CommerceService, o domain.OrderRepository) *CommerceHandler {
	return &CommerceHandler{svc: s, orders: o}
}

type bannerRequest struct {
	Title           string `json:"title" validate:"max=160"`
	Subtitle        string `json:"subtitle" validate:"max=300"`
	DesktopImageURL string `json:"desktop_image_url" validate:"required,max=1000"`
	MobileImageURL  string `json:"mobile_image_url" validate:"required,max=1000"`
	LinkURL         string `json:"link_url" validate:"max=1000"`
	ButtonLabel     string `json:"button_label" validate:"max=50"`
	SortOrder       int    `json:"sort_order"`
	IsActive        bool   `json:"is_active"`
}
type bannerResponse struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Subtitle        string    `json:"subtitle"`
	DesktopImageURL string    `json:"desktop_image_url"`
	MobileImageURL  string    `json:"mobile_image_url"`
	LinkURL         string    `json:"link_url"`
	ButtonLabel     string    `json:"button_label"`
	SortOrder       int       `json:"sort_order"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Keep commerce API responses explicit: domain structs are persistence/service
// models and their Go field names are not the JSON contract consumed by React.
type shippingMethodResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Province    string    `json:"province"`
	City        string    `json:"city"`
	Price       int64     `json:"price"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func shippingMethodDTO(m *domain.ShippingMethod) shippingMethodResponse {
	return shippingMethodResponse{m.ID, m.Name, m.Description, m.Province, m.City, m.Price, m.IsActive, m.SortOrder, m.CreatedAt, m.UpdatedAt}
}

type couponResponse struct {
	ID              int64      `json:"id"`
	Code            string     `json:"code"`
	Kind            string     `json:"kind"`
	Value           int64      `json:"value"`
	MinimumSubtotal int64      `json:"minimum_subtotal"`
	MaximumDiscount *int64     `json:"maximum_discount"`
	UsageLimit      *int64     `json:"usage_limit"`
	UsedCount       int64      `json:"used_count"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func couponDTO(c *domain.Coupon) couponResponse {
	return couponResponse{c.ID, c.Code, c.Kind, c.Value, c.MinimumSubtotal, c.MaximumDiscount, c.UsageLimit, c.UsedCount, c.StartsAt, c.EndsAt, c.IsActive, c.CreatedAt, c.UpdatedAt}
}

func bannerDTO(b *domain.Banner) bannerResponse {
	return bannerResponse{b.ID, b.Title, b.Subtitle, b.DesktopImageURL, b.MobileImageURL, b.LinkURL, b.ButtonLabel, b.SortOrder, b.IsActive, b.CreatedAt, b.UpdatedAt}
}
func (h *CommerceHandler) ListBanners(c echo.Context) error {
	items, err := h.svc.Banners(c.Request().Context(), false)
	if err != nil {
		return err
	}
	out := make([]bannerResponse, 0, len(items))
	for i := range items {
		out = append(out, bannerDTO(&items[i]))
	}
	return c.JSON(http.StatusOK, out)
}
func (h *CommerceHandler) ListAdminBanners(c echo.Context) error {
	items, err := h.svc.Banners(c.Request().Context(), true)
	if err != nil {
		return err
	}
	out := make([]bannerResponse, 0, len(items))
	for i := range items {
		out = append(out, bannerDTO(&items[i]))
	}
	return c.JSON(http.StatusOK, out)
}
func (h *CommerceHandler) SaveBanner(c echo.Context) error {
	var req bannerRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	var id int64
	if c.Param("id") != "" {
		n, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return domain.ErrNotFound
		}
		id = n
	}
	b, err := h.svc.SaveBanner(c.Request().Context(), &domain.Banner{ID: id, Title: req.Title, Subtitle: req.Subtitle, DesktopImageURL: req.DesktopImageURL, MobileImageURL: req.MobileImageURL, LinkURL: req.LinkURL, ButtonLabel: req.ButtonLabel, SortOrder: req.SortOrder, IsActive: req.IsActive})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, bannerDTO(b))
}
func (h *CommerceHandler) DeleteBanner(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err = h.svc.DeleteBanner(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type shippingRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=300"`
	Province    string `json:"province" validate:"max=100"`
	City        string `json:"city" validate:"max=100"`
	Price       *int64 `json:"price" validate:"required,gte=0"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (h *CommerceHandler) ListShipping(c echo.Context) error {
	items, err := h.svc.ShippingMethods(c.Request().Context(), c.QueryParam("province"), c.QueryParam("city"), false)
	if err != nil {
		return err
	}
	out := make([]shippingMethodResponse, 0, len(items))
	for i := range items {
		out = append(out, shippingMethodDTO(&items[i]))
	}
	return c.JSON(http.StatusOK, out)
}
func (h *CommerceHandler) ListAdminShipping(c echo.Context) error {
	items, err := h.svc.ShippingMethods(c.Request().Context(), "", "", true)
	if err != nil {
		return err
	}
	out := make([]shippingMethodResponse, 0, len(items))
	for i := range items {
		out = append(out, shippingMethodDTO(&items[i]))
	}
	return c.JSON(http.StatusOK, out)
}
func (h *CommerceHandler) SaveShipping(c echo.Context) error {
	var req shippingRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	var id int64
	if c.Param("id") != "" {
		n, e := strconv.ParseInt(c.Param("id"), 10, 64)
		if e != nil {
			return domain.ErrNotFound
		}
		id = n
	}
	m, err := h.svc.SaveShipping(c.Request().Context(), &domain.ShippingMethod{ID: id, Name: req.Name, Description: req.Description, Province: req.Province, City: req.City, Price: *req.Price, IsActive: req.IsActive, SortOrder: req.SortOrder})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, shippingMethodDTO(m))
}
func (h *CommerceHandler) DeleteShipping(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err = h.svc.DeleteShipping(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type couponRequest struct {
	Code            string     `json:"code" validate:"required,max=50"`
	Kind            string     `json:"kind" validate:"required,oneof=percent fixed"`
	Value           int64      `json:"value" validate:"required,gt=0"`
	MinimumSubtotal int64      `json:"minimum_subtotal" validate:"gte=0"`
	MaximumDiscount *int64     `json:"maximum_discount"`
	UsageLimit      *int64     `json:"usage_limit"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	IsActive        bool       `json:"is_active"`
}

func (h *CommerceHandler) ListCoupons(c echo.Context) error {
	items, err := h.svc.Coupons(c.Request().Context())
	if err != nil {
		return err
	}
	out := make([]couponResponse, 0, len(items))
	for i := range items {
		out = append(out, couponDTO(&items[i]))
	}
	return c.JSON(http.StatusOK, out)
}
func (h *CommerceHandler) SaveCoupon(c echo.Context) error {
	var req couponRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	var id int64
	if c.Param("id") != "" {
		n, e := strconv.ParseInt(c.Param("id"), 10, 64)
		if e != nil {
			return domain.ErrNotFound
		}
		id = n
	}
	coupon, err := h.svc.SaveCoupon(c.Request().Context(), &domain.Coupon{ID: id, Code: req.Code, Kind: req.Kind, Value: req.Value, MinimumSubtotal: req.MinimumSubtotal, MaximumDiscount: req.MaximumDiscount, UsageLimit: req.UsageLimit, StartsAt: req.StartsAt, EndsAt: req.EndsAt, IsActive: req.IsActive})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, couponDTO(coupon))
}
func (h *CommerceHandler) DeleteCoupon(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err = h.svc.DeleteCoupon(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *CommerceHandler) ValidateCoupon(c echo.Context) error {
	var req struct {
		Code     string `json:"code" validate:"required,max=50"`
		Subtotal int64  `json:"subtotal" validate:"required,gt=0"`
	}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	coupon, discount, err := h.svc.PreviewCoupon(c.Request().Context(), req.Code, req.Subtotal)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"code": coupon.Code, "discount_amount": discount})
}

type TrackingResponse struct {
	TrackingCode       string             `json:"tracking_code"`
	Status             domain.OrderStatus `json:"status"`
	ShippingMethodName string             `json:"shipping_method_name"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

func (h *CommerceHandler) Track(c echo.Context) error {
	code := strings.ToUpper(strings.TrimSpace(c.Param("code")))
	if len(code) < 8 || len(code) > 40 {
		return domain.ErrNotFound
	}
	order, err := h.orders.GetByTrackingCode(c.Request().Context(), code)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, TrackingResponse{order.TrackingCode, order.Status, order.ShippingMethodName, order.CreatedAt, order.UpdatedAt})
}
