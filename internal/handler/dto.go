package handler

import (
	"math/big"
	"time"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

// ---------- auth (phone + OTP) ----------

type OTPRequestRequest struct {
	Phone string `json:"phone" validate:"required"`
}

type OTPRequestResponse struct {
	Message          string `json:"message"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type OTPVerifyRequest struct {
	Phone     string `json:"phone" validate:"required"`
	Code      string `json:"code" validate:"required"`
	Username  string `json:"username" validate:"omitempty,max=30"`
	FirstName string `json:"first_name" validate:"omitempty,max=100"`
	LastName  string `json:"last_name" validate:"omitempty,max=100"`
	Email     string `json:"email" validate:"omitempty,email,max=254"`
}

// ---------- users ----------

type UserResponse struct {
	ID        int64       `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Username  string      `json:"username"`
	Phone     string      `json:"phone"`
	Email     *string     `json:"email,omitempty"`
	Role      domain.Role `json:"role"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        UserResponse `json:"user"`
}

// ---------- addresses ----------

type AddressRequest struct {
	ReceiverName  string `json:"receiver_name" validate:"required,max=100"`
	ReceiverPhone string `json:"receiver_phone" validate:"required"`
	Province      string `json:"province" validate:"required,max=100"`
	City          string `json:"city" validate:"required,max=100"`
	AddressLine   string `json:"address_line" validate:"required,max=500"`
	PostalCode    string `json:"postal_code" validate:"required"`
}

type AddressResponse struct {
	ID            int64     `json:"id"`
	ReceiverName  string    `json:"receiver_name"`
	ReceiverPhone string    `json:"receiver_phone"`
	Province      string    `json:"province"`
	City          string    `json:"city"`
	AddressLine   string    `json:"address_line"`
	PostalCode    string    `json:"postal_code"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ---------- products ----------

type ProductRequest struct {
	Name          string                  `json:"name" validate:"required,max=200"`
	Slug          string                  `json:"slug" validate:"omitempty,max=180"`
	Description   string                  `json:"description" validate:"max=100000"`
	ImageURL      string                  `json:"image_url" validate:"omitempty,max=1000"`
	ImageURLs     []string                `json:"image_urls" validate:"max=8,dive,max=1000"`
	Attributes    map[string]string       `json:"attributes"`
	VariantLabel  string                  `json:"variant_label" validate:"omitempty,max=50"`
	Variants      []domain.ProductVariant `json:"variants" validate:"max=20,dive"`
	CategoryID    *int64                  `json:"category_id" validate:"required,gt=0"`
	BrandID       int64                   `json:"brand_id" validate:"required,gt=0"`
	Price         int64                   `json:"price" validate:"gte=0"`
	DiscountPrice int64                   `json:"discount_price" validate:"gte=0"`
	Stock         int                     `json:"stock" validate:"gte=0"`
	IsActive      *bool                   `json:"is_active"`
	IsPopular     bool                    `json:"is_popular"`
}

type ProductResponse struct {
	ID              int64                    `json:"id"`
	Name            string                   `json:"name"`
	Slug            string                   `json:"slug"`
	Description     string                   `json:"description"`
	ImageURL        string                   `json:"image_url,omitempty"`
	ImageURLs       []string                 `json:"image_urls"`
	Attributes      map[string]string        `json:"attributes"`
	VariantLabel    string                   `json:"variant_label,omitempty"`
	Variants        []ProductVariantResponse `json:"variants"`
	CategoryID      *int64                   `json:"category_id,omitempty"`
	CategoryName    *string                  `json:"category_name,omitempty"`
	CategorySlug    string                   `json:"category_slug"`
	BrandID         int64                    `json:"brand_id"`
	BrandName       string                   `json:"brand_name"`
	BrandSlug       string                   `json:"brand_slug"`
	Price           int64                    `json:"price"`
	DiscountPrice   int64                    `json:"discount_price"`
	DiscountPercent int                      `json:"discount_percent"`
	Stock           int                      `json:"stock"`
	IsActive        bool                     `json:"is_active"`
	IsPopular       bool                     `json:"is_popular"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

type ProductVariantResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ImageURL        string `json:"image_url,omitempty"`
	Price           int64  `json:"price"`
	DiscountPrice   int64  `json:"discount_price"`
	DiscountPercent int    `json:"discount_percent"`
	Stock           int    `json:"stock"`
}

type ProductVisibilityRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

// ---------- orders ----------

type OrderItemRequest struct {
	ProductID   int64  `json:"product_id" validate:"required,gt=0"`
	VariantID   string `json:"variant_id" validate:"omitempty,max=80"`
	VariantName string `json:"variant_name" validate:"omitempty,max=80"`
	Quantity    int    `json:"quantity" validate:"required,gt=0,max=1000"`
}

type PlaceOrderRequest struct {
	AddressID        int64              `json:"address_id" validate:"required,gt=0"`
	ShippingMethodID int64              `json:"shipping_method_id" validate:"required,gt=0"`
	CouponCode       string             `json:"coupon_code" validate:"omitempty,max=50"`
	Items            []OrderItemRequest `json:"items" validate:"required,min=1,max=50,dive"`
}

type OrderItemResponse struct {
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	VariantID    string `json:"variant_id,omitempty"`
	VariantLabel string `json:"variant_label,omitempty"`
	VariantName  string `json:"variant_name,omitempty"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	Subtotal     int64  `json:"subtotal"`
}

type ShippingAddressResponse struct {
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`
	Province      string `json:"province"`
	City          string `json:"city"`
	AddressLine   string `json:"address_line"`
	PostalCode    string `json:"postal_code"`
}

type OrderPaymentResponse struct {
	Authority string               `json:"authority"`
	Status    domain.PaymentStatus `json:"status"`
	RefID     *string              `json:"ref_id,omitempty"`
	PaidAt    *time.Time           `json:"paid_at,omitempty"`
}

type OrderResponse struct {
	ID                 int64                   `json:"id"`
	UserID             int64                   `json:"user_id"`
	CustomerName       string                  `json:"customer_name,omitempty"`
	CustomerPhone      string                  `json:"customer_phone,omitempty"`
	Status             domain.OrderStatus      `json:"status"`
	TotalAmount        int64                   `json:"total_amount"`
	ItemsAmount        int64                   `json:"items_amount"`
	DiscountAmount     int64                   `json:"discount_amount"`
	ShippingCost       int64                   `json:"shipping_cost"`
	CouponCode         string                  `json:"coupon_code,omitempty"`
	ShippingMethodName string                  `json:"shipping_method_name"`
	TrackingCode       string                  `json:"tracking_code"`
	ShippingAddress    ShippingAddressResponse `json:"shipping_address"`
	Items              []OrderItemResponse     `json:"items"`
	Payment            *OrderPaymentResponse   `json:"payment,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// ---------- product comments ----------

type ProductCommentRequest struct {
	Body string `json:"body" validate:"required,max=2000"`
}

type CommentApprovalRequest struct {
	IsApproved *bool `json:"is_approved" validate:"required"`
}

type ProductCommentResponse struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	ProductSlug string    `json:"product_slug,omitempty"`
	AuthorName  string    `json:"author_name"`
	Body        string    `json:"body"`
	IsApproved  bool      `json:"is_approved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryRequest struct {
	Name           string `json:"name" validate:"required,max=120"`
	Slug           string `json:"slug" validate:"omitempty,max=180"`
	Description    string `json:"description" validate:"max=100000"`
	SEOTitle       string `json:"seo_title" validate:"omitempty,max=200"`
	SEODescription string `json:"seo_description" validate:"omitempty,max=320"`
	HomeTitle      string `json:"home_title" validate:"omitempty,max=120"`
	HomeImageURL   string `json:"home_image_url" validate:"omitempty,max=1000"`
	ShowOnHome     bool   `json:"show_on_home"`
	ParentID       *int64 `json:"parent_id" validate:"omitempty,gt=0"`
}

type CategoryResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	SEOTitle       string    `json:"seo_title"`
	SEODescription string    `json:"seo_description"`
	HomeTitle      string    `json:"home_title"`
	HomeImageURL   string    `json:"home_image_url"`
	ShowOnHome     bool      `json:"show_on_home"`
	ParentID       *int64    `json:"parent_id,omitempty"`
	ParentName     *string   `json:"parent_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BrandRequest struct {
	Name           string `json:"name" validate:"required,max=120"`
	Slug           string `json:"slug" validate:"omitempty,max=180"`
	Description    string `json:"description" validate:"max=100000"`
	SEOTitle       string `json:"seo_title" validate:"omitempty,max=200"`
	SEODescription string `json:"seo_description" validate:"omitempty,max=320"`
}

type BrandResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	SEOTitle       string    `json:"seo_title"`
	SEODescription string    `json:"seo_description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ChangeUserAccessRequest struct {
	Role     domain.Role `json:"role" validate:"required,oneof=user admin"`
	IsActive *bool       `json:"is_active" validate:"required"`
}

type ChangeOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status" validate:"required,oneof=processing shipped cancelled"`
}

type AdminOverviewResponse struct {
	Products      int64 `json:"products"`
	Categories    int64 `json:"categories"`
	Users         int64 `json:"users"`
	Orders        int64 `json:"orders"`
	PendingOrders int64 `json:"pending_orders"`
	PaidRevenue   int64 `json:"paid_revenue"`
}

// ---------- payments ----------

type ConfirmPaymentRequest struct {
	Action string `json:"action" validate:"required,oneof=pay cancel"`
}

type PaymentGatewayResponse struct {
	Authority string               `json:"authority"`
	OrderID   int64                `json:"order_id"`
	Amount    int64                `json:"amount"`
	Status    domain.PaymentStatus `json:"status"`
	RefID     *string              `json:"ref_id,omitempty"`
}

// ---------- pagination ----------

// PageResponse wraps a paginated list.
type PageResponse[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// ---------- mappers ----------

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Username: u.Username, Phone: u.Phone,
		Email: u.Email, Role: u.Role, IsActive: u.IsActive, CreatedAt: u.CreatedAt,
	}
}

func toAddressResponse(a *domain.Address) AddressResponse {
	return AddressResponse{
		ID: a.ID, ReceiverName: a.ReceiverName, ReceiverPhone: a.ReceiverPhone,
		Province: a.Province, City: a.City, AddressLine: a.AddressLine, PostalCode: a.PostalCode,
		IsDefault: a.IsDefault, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func toProductResponse(p *domain.Product) ProductResponse {
	variants := make([]ProductVariantResponse, 0, len(p.Variants))
	for _, variant := range p.Variants {
		variants = append(variants, ProductVariantResponse{
			ID: variant.ID, Name: variant.Name, ImageURL: variant.ImageURL, Price: variant.Price,
			DiscountPrice: variant.DiscountPrice, DiscountPercent: discountPercent(variant.Price, variant.DiscountPrice), Stock: variant.Stock,
		})
	}
	return ProductResponse{
		ID:              p.ID,
		Name:            p.Name,
		Slug:            p.Slug,
		Description:     richtext.Sanitize(p.Description),
		ImageURL:        p.ImageURL,
		ImageURLs:       p.ImageURLs,
		Attributes:      p.Attributes,
		VariantLabel:    p.VariantLabel,
		Variants:        variants,
		CategoryID:      p.CategoryID,
		CategoryName:    p.CategoryName,
		CategorySlug:    p.CategorySlug,
		BrandID:         p.BrandID,
		BrandName:       p.BrandName,
		BrandSlug:       p.BrandSlug,
		Price:           p.Price,
		DiscountPrice:   p.DiscountPrice,
		DiscountPercent: discountPercent(p.Price, p.DiscountPrice),
		Stock:           p.Stock,
		IsActive:        p.IsActive,
		IsPopular:       p.IsPopular,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func discountPercent(price, discountPrice int64) int {
	if price <= 0 || discountPrice <= 0 || discountPrice >= price {
		return 0
	}
	difference := new(big.Int).Sub(big.NewInt(price), big.NewInt(discountPrice))
	difference.Mul(difference, big.NewInt(100))
	difference.Div(difference, big.NewInt(price))
	return int(difference.Int64())
}

func toProductPage(products []domain.Product, page, limit, total int) PageResponse[ProductResponse] {
	data := make([]ProductResponse, 0, len(products))
	for i := range products {
		data = append(data, toProductResponse(&products[i]))
	}
	return PageResponse[ProductResponse]{Data: data, Page: page, Limit: limit, Total: total}
}

func toOrderResponse(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID:    it.ProductID,
			ProductName:  it.ProductName,
			VariantID:    it.VariantID,
			VariantLabel: it.VariantLabel,
			VariantName:  it.VariantName,
			Quantity:     it.Quantity,
			UnitPrice:    it.UnitPrice,
			Subtotal:     it.UnitPrice * int64(it.Quantity),
		})
	}
	resp := OrderResponse{
		ID: o.ID, UserID: o.UserID, CustomerName: o.CustomerName, CustomerPhone: o.CustomerPhone,
		Status: o.Status, TotalAmount: o.TotalAmount, ItemsAmount: o.ItemsAmount,
		DiscountAmount: o.DiscountAmount, ShippingCost: o.ShippingCost, CouponCode: o.CouponCode,
		ShippingMethodName: o.ShippingMethodName, TrackingCode: o.TrackingCode,
		ShippingAddress: ShippingAddressResponse{
			ReceiverName: o.ShippingReceiverName, ReceiverPhone: o.ShippingReceiverPhone,
			Province: o.ShippingProvince, City: o.ShippingCity,
			AddressLine: o.ShippingAddressLine, PostalCode: o.ShippingPostalCode,
		},
		Items: items, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
	if o.Payment != nil {
		resp.Payment = &OrderPaymentResponse{
			Authority: o.Payment.Authority, Status: o.Payment.Status, RefID: o.Payment.RefID, PaidAt: o.Payment.PaidAt,
		}
	}
	return resp
}

func toOrderPage(orders []domain.Order, page, limit, total int) PageResponse[OrderResponse] {
	data := make([]OrderResponse, 0, len(orders))
	for i := range orders {
		data = append(data, toOrderResponse(&orders[i]))
	}
	return PageResponse[OrderResponse]{Data: data, Page: page, Limit: limit, Total: total}
}

func toPaymentGatewayResponse(p *domain.Payment) PaymentGatewayResponse {
	return PaymentGatewayResponse{Authority: p.Authority, OrderID: p.OrderID, Amount: p.Amount, Status: p.Status, RefID: p.RefID}
}

func toProductCommentResponse(comment *domain.ProductComment) ProductCommentResponse {
	return ProductCommentResponse{ID: comment.ID, ProductID: comment.ProductID, ProductName: comment.ProductName, ProductSlug: comment.ProductSlug,
		AuthorName: comment.AuthorName, Body: comment.Body, IsApproved: comment.IsApproved,
		CreatedAt: comment.CreatedAt, UpdatedAt: comment.UpdatedAt}
}

func toProductCommentPage(comments []domain.ProductComment, page, limit, total int) PageResponse[ProductCommentResponse] {
	data := make([]ProductCommentResponse, 0, len(comments))
	for i := range comments {
		data = append(data, toProductCommentResponse(&comments[i]))
	}
	return PageResponse[ProductCommentResponse]{Data: data, Page: page, Limit: limit, Total: total}
}
