package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"shop-api/internal/domain"
	"shop-api/internal/richtext"
	"shop-api/internal/service"
)

type BrandHandler struct{ svc *service.BrandService }

func NewBrandHandler(svc *service.BrandService) *BrandHandler { return &BrandHandler{svc: svc} }

func (h *BrandHandler) List(c echo.Context) error {
	brands, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]BrandResponse, 0, len(brands))
	for i := range brands {
		data = append(data, toBrandResponse(&brands[i]))
	}
	return c.JSON(http.StatusOK, data)
}
func (h *BrandHandler) GetBySlug(c echo.Context) error {
	brand, err := h.svc.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBrandResponse(brand))
}
func (h *BrandHandler) Create(c echo.Context) error {
	var req BrandRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	brand, err := h.svc.Create(c.Request().Context(), service.BrandInput{Name: req.Name, Slug: req.Slug, Description: req.Description, SEOTitle: req.SEOTitle, SEODescription: req.SEODescription})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toBrandResponse(brand))
}
func (h *BrandHandler) Update(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req BrandRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	brand, err := h.svc.Update(c.Request().Context(), id, service.BrandInput{Name: req.Name, Slug: req.Slug, Description: req.Description, SEOTitle: req.SEOTitle, SEODescription: req.SEODescription})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toBrandResponse(brand))
}
func (h *BrandHandler) Delete(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
func toBrandResponse(brand *domain.Brand) BrandResponse {
	return BrandResponse{ID: brand.ID, Name: brand.Name, Slug: brand.Slug, Description: richtext.Sanitize(brand.Description), SEOTitle: brand.SEOTitle, SEODescription: brand.SEODescription, CreatedAt: brand.CreatedAt, UpdatedAt: brand.UpdatedAt}
}
