package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create handles POST /api/v1/orders: the authenticated user buys products,
// shipping to one of their saved addresses.
func (h *OrderHandler) Create(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}

	var req PlaceOrderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	items := make([]domain.PurchaseItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, domain.PurchaseItem{ProductID: it.ProductID, VariantID: it.VariantID, VariantName: it.VariantName, Quantity: it.Quantity})
	}

	order, err := h.svc.PlaceOrder(c.Request().Context(), userID, req.AddressID, req.ShippingMethodID, req.CouponCode, items)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toOrderResponse(order))
}

// List handles GET /api/v1/orders: the user's own orders.
func (h *OrderHandler) List(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	page, limit := pagination(c)

	orders, total, err := h.svc.ListByUser(c.Request().Context(), userID, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toOrderPage(orders, page, limit, total))
}

// Get handles GET /api/v1/orders/:id (only the owner can see an order).
func (h *OrderHandler) Get(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	id, err := idParam(c)
	if err != nil {
		return err
	}

	order, err := h.svc.Get(c.Request().Context(), userID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toOrderResponse(order))
}
