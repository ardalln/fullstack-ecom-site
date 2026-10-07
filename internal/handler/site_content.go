package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type SiteContentHandler struct{ service *service.SiteContentService }

func NewSiteContentHandler(contentService *service.SiteContentService) *SiteContentHandler {
	return &SiteContentHandler{service: contentService}
}

type SiteContentRequest struct {
	ContactPhone   string `json:"contact_phone" validate:"max=40"`
	ContactEmail   string `json:"contact_email" validate:"max=254"`
	ContactAddress string `json:"contact_address" validate:"max=500"`
	ContactHours   string `json:"contact_hours" validate:"max=120"`
	InstagramURL   string `json:"instagram_url" validate:"max=500"`
	AboutContent   string `json:"about_content" validate:"max=200000"`
	TermsContent   string `json:"terms_content" validate:"max=200000"`
}

type SiteContentResponse struct {
	ContactPhone   string `json:"contact_phone"`
	ContactEmail   string `json:"contact_email"`
	ContactAddress string `json:"contact_address"`
	ContactHours   string `json:"contact_hours"`
	InstagramURL   string `json:"instagram_url"`
	AboutContent   string `json:"about_content"`
	TermsContent   string `json:"terms_content"`
}

func (h *SiteContentHandler) Get(c echo.Context) error {
	content, err := h.service.Get(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toSiteContentResponse(content))
}

func (h *SiteContentHandler) Update(c echo.Context) error {
	var request SiteContentRequest
	if err := bindAndValidate(c, &request); err != nil {
		return err
	}
	content, err := h.service.Update(c.Request().Context(), service.SiteContentInput{
		ContactPhone: request.ContactPhone, ContactEmail: request.ContactEmail,
		ContactAddress: request.ContactAddress, ContactHours: request.ContactHours,
		InstagramURL: request.InstagramURL, AboutContent: request.AboutContent, TermsContent: request.TermsContent,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toSiteContentResponse(content))
}

func toSiteContentResponse(content *domain.SiteContent) SiteContentResponse {
	return SiteContentResponse{
		ContactPhone: content.ContactPhone, ContactEmail: content.ContactEmail,
		ContactAddress: content.ContactAddress, ContactHours: content.ContactHours,
		InstagramURL: content.InstagramURL, AboutContent: content.AboutContent, TermsContent: content.TermsContent,
	}
}
