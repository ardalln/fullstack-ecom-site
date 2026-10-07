package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type AdminHandler struct{ svc *service.AdminService }

func NewAdminHandler(svc *service.AdminService) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) Overview(c echo.Context) error {
	stats, err := h.svc.Overview(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, AdminOverviewResponse{
		Products: stats.Products, Categories: stats.Categories, Users: stats.Users,
		Orders: stats.Orders, PendingOrders: stats.PendingOrders, PaidRevenue: stats.PaidRevenue,
	})
}

func (h *AdminHandler) ListUsers(c echo.Context) error {
	page, limit := pagination(c)
	search := strings.TrimSpace(c.QueryParam("q"))
	if len(search) > 100 {
		return fmt.Errorf("%w: search query must be at most 100 characters", domain.ErrInvalidInput)
	}
	users, total, err := h.svc.ListUsers(c.Request().Context(), search, page, limit)
	if err != nil {
		return err
	}
	data := make([]UserResponse, 0, len(users))
	for i := range users {
		data = append(data, toUserResponse(&users[i]))
	}
	return c.JSON(http.StatusOK, PageResponse[UserResponse]{Data: data, Page: page, Limit: limit, Total: total})
}

func (h *AdminHandler) ChangeUserAccess(c echo.Context) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return err
	}
	targetID, err := idParam(c)
	if err != nil {
		return err
	}
	var req ChangeUserAccessRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	user, err := h.svc.ChangeUserAccess(c.Request().Context(), actorID, targetID, req.Role, *req.IsActive)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *AdminHandler) ListOrders(c echo.Context) error {
	page, limit := pagination(c)
	status := domain.OrderStatus(strings.TrimSpace(c.QueryParam("status")))
	if !validOrderFilter(status) {
		return fmt.Errorf("%w: unsupported order status filter", domain.ErrInvalidInput)
	}
	orders, total, err := h.svc.ListOrders(c.Request().Context(), status, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toOrderPage(orders, page, limit, total))
}

func (h *AdminHandler) ChangeOrderStatus(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req ChangeOrderStatusRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	order, err := h.svc.ChangeOrderStatus(c.Request().Context(), id, req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toOrderResponse(order))
}

func validOrderFilter(status domain.OrderStatus) bool {
	switch status {
	case "", domain.OrderStatusAwaitingPayment, domain.OrderStatusPaid, domain.OrderStatusProcessing,
		domain.OrderStatusShipped, domain.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
