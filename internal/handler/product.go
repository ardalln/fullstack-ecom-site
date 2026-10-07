package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (r ProductRequest) toInput() service.ProductInput {
	return service.ProductInput{
		Name:          r.Name,
		Slug:          r.Slug,
		Description:   r.Description,
		ImageURL:      r.ImageURL,
		ImageURLs:     r.ImageURLs,
		Attributes:    r.Attributes,
		VariantLabel:  r.VariantLabel,
		Variants:      r.Variants,
		CategoryID:    r.CategoryID,
		BrandID:       r.BrandID,
		Price:         r.Price,
		DiscountPrice: r.DiscountPrice,
		Stock:         r.Stock,
		IsActive:      r.IsActive,
		IsPopular:     r.IsPopular,
	}
}

// List handles GET /api/v1/products?page=&limit=.
func (h *ProductHandler) List(c echo.Context) error {
	return h.list(c, false)
}

func (h *ProductHandler) ListAdmin(c echo.Context) error {
	return h.list(c, true)
}

func (h *ProductHandler) list(c echo.Context, admin bool) error {
	page, limit := pagination(c)
	var categoryID, brandID *int64
	search := strings.TrimSpace(c.QueryParam("q"))
	if len(search) > 100 {
		return fmt.Errorf("%w: search query must be at most 100 characters", domain.ErrInvalidInput)
	}
	if rawCategory := c.QueryParam("category_id"); rawCategory != "" {
		parsed, parseErr := strconv.ParseInt(rawCategory, 10, 64)
		if parseErr != nil || parsed <= 0 {
			return fmt.Errorf("%w: category_id must be a positive integer", domain.ErrInvalidInput)
		}
		categoryID = &parsed
	}
	if rawBrand := c.QueryParam("brand_id"); rawBrand != "" {
		parsed, parseErr := strconv.ParseInt(rawBrand, 10, 64)
		if parseErr != nil || parsed <= 0 {
			return fmt.Errorf("%w: brand_id must be a positive integer", domain.ErrInvalidInput)
		}
		brandID = &parsed
	}
	var products []domain.Product
	var total int
	var err error
	if admin {
		products, total, err = h.svc.ListAdmin(c.Request().Context(), search, categoryID, brandID, page, limit)
	} else {
		products, total, err = h.svc.List(c.Request().Context(), search, categoryID, brandID, page, limit)
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toProductPage(products, page, limit, total))
}

func (h *ProductHandler) GetBySlug(c echo.Context) error {
	product, err := h.svc.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toProductResponse(product))
}

// Get handles GET /api/v1/products/:id.
func (h *ProductHandler) Get(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	product, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toProductResponse(product))
}

// Create handles POST /api/v1/products (admin only).
func (h *ProductHandler) Create(c echo.Context) error {
	var req ProductRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	product, err := h.svc.Create(c.Request().Context(), req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toProductResponse(product))
}

// Update handles PUT /api/v1/products/:id (admin only, full replacement).
func (h *ProductHandler) Update(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req ProductRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	product, err := h.svc.Update(c.Request().Context(), id, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toProductResponse(product))
}

// Delete handles DELETE /api/v1/products/:id (admin only).
func (h *ProductHandler) Delete(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ProductHandler) SetVisibility(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req ProductVisibilityRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetActive(c.Request().Context(), id, *req.IsActive); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
