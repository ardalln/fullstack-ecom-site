package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/service"
)

type AddressHandler struct {
	svc *service.AddressService
}

func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

func (r AddressRequest) toInput() service.AddressInput {
	return service.AddressInput{
		ReceiverName: r.ReceiverName, ReceiverPhone: r.ReceiverPhone,
		Province: r.Province, City: r.City, AddressLine: r.AddressLine, PostalCode: r.PostalCode,
	}
}

// List handles GET /api/v1/addresses.
func (h *AddressHandler) List(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	addresses, err := h.svc.List(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	out := make([]AddressResponse, 0, len(addresses))
	for i := range addresses {
		out = append(out, toAddressResponse(&addresses[i]))
	}
	return c.JSON(http.StatusOK, out)
}

// Create handles POST /api/v1/addresses.
func (h *AddressHandler) Create(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	var req AddressRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	address, err := h.svc.Create(c.Request().Context(), userID, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toAddressResponse(address))
}

// Update handles PUT /api/v1/addresses/:id.
func (h *AddressHandler) Update(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req AddressRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	address, err := h.svc.Update(c.Request().Context(), userID, id, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toAddressResponse(address))
}

// Delete handles DELETE /api/v1/addresses/:id.
func (h *AddressHandler) Delete(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), userID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// SetDefault handles PATCH /api/v1/addresses/:id/default.
func (h *AddressHandler) SetDefault(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	id, err := idParam(c)
	if err != nil {
		return err
	}
	if err := h.svc.SetDefault(c.Request().Context(), userID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
