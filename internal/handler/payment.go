package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// Request handles POST /api/v1/orders/:id/pay (order owner only). Starts (or
// resumes) a checkout session and returns its authority token.
func (h *PaymentHandler) Request(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	orderID, err := idParam(c)
	if err != nil {
		return err
	}

	payment, err := h.svc.RequestPayment(c.Request().Context(), userID, orderID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toPaymentGatewayResponse(payment))
}

// Gateway handles GET /api/v1/payments/:authority. Intentionally public: the
// random authority token is the capability, exactly like a real payment
// gateway's checkout link needs no separate login.
func (h *PaymentHandler) Gateway(c echo.Context) error {
	payment, err := h.svc.GatewayInfo(c.Request().Context(), c.Param("authority"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toPaymentGatewayResponse(payment))
}

// Confirm handles POST /api/v1/payments/:authority/confirm. Also public,
// simulating the bank redirecting back with the result of the payment.
func (h *PaymentHandler) Confirm(c echo.Context) error {
	var req ConfirmPaymentRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	payment, err := h.svc.ConfirmPayment(c.Request().Context(), c.Param("authority"), req.Action)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toPaymentGatewayResponse(payment))
}
