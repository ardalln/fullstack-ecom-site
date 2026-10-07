package domain

import "time"

// Product is an item that can be purchased.
// Price is stored in the smallest currency unit (e.g. cents) to avoid float errors.
type Product struct {
	ID            int64
	Name          string
	Slug          string
	Description   string
	ImageURL      string
	ImageURLs     []string
	Attributes    map[string]string
	VariantLabel  string
	Variants      []ProductVariant
	CategoryID    *int64
	CategoryName  *string
	CategorySlug  string
	BrandID       int64
	BrandName     string
	BrandSlug     string
	Price         int64
	DiscountPrice int64
	Stock         int
	IsActive      bool
	IsPopular     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProductVariant is a selectable option for a product, such as a color or size.
type ProductVariant struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ImageURL      string `json:"image_url,omitempty"`
	Price         int64  `json:"price"`
	DiscountPrice int64  `json:"discount_price,omitempty"`
	Stock         int    `json:"stock"`
}
