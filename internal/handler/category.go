package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
	"shop-api/internal/service"
)

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c echo.Context) error {
	categories, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		data = append(data, toCategoryResponse(&category))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *CategoryHandler) GetBySlug(c echo.Context) error {
	category, err := h.svc.GetBySlug(c.Request().Context(), c.Param("slug"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toCategoryResponse(category))
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req CategoryRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	category, err := h.svc.Create(c.Request().Context(), service.CategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, SEOTitle: req.SEOTitle, SEODescription: req.SEODescription, HomeTitle: req.HomeTitle, HomeImageURL: req.HomeImageURL, ShowOnHome: req.ShowOnHome, ParentID: req.ParentID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toCategoryResponse(category))
}

func (h *CategoryHandler) Update(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req CategoryRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	category, err := h.svc.Update(c.Request().Context(), id, service.CategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, SEOTitle: req.SEOTitle, SEODescription: req.SEODescription, HomeTitle: req.HomeTitle, HomeImageURL: req.HomeImageURL, ShowOnHome: req.ShowOnHome, ParentID: req.ParentID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toCategoryResponse(category))
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func toCategoryResponse(category *domain.Category) CategoryResponse {
	return CategoryResponse{ID: category.ID, Name: category.Name, Slug: category.Slug, Description: richtext.Sanitize(category.Description), SEOTitle: category.SEOTitle, SEODescription: category.SEODescription, HomeTitle: category.HomeTitle, HomeImageURL: category.HomeImageURL, ShowOnHome: category.ShowOnHome,
		ParentID: category.ParentID, ParentName: category.ParentName,
		CreatedAt: category.CreatedAt, UpdatedAt: category.UpdatedAt}
}
