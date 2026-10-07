package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"shop-api/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RequestOTP handles POST /api/v1/auth/otp/request.
func (h *AuthHandler) RequestOTP(c echo.Context) error {
	var req OTPRequestRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ttl, err := h.svc.RequestOTP(c.Request().Context(), req.Phone)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, OTPRequestResponse{Message: "کد تایید ارسال شد.", ExpiresInSeconds: int(ttl.Seconds())})
}

// VerifyOTP handles POST /api/v1/auth/otp/verify. Logs an existing user in,
// or creates a new account (first_name/last_name required for new users).
func (h *AuthHandler) VerifyOTP(c echo.Context) error {
	var req OTPVerifyRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.svc.VerifyOTP(c.Request().Context(), service.VerifyOTPInput{
		Phone: req.Phone, Code: req.Code, Username: req.Username, FirstName: req.FirstName, LastName: req.LastName, Email: req.Email,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   result.ExpiresAt,
		User:        toUserResponse(result.User),
	})
}

// Me handles GET /api/v1/me.
func (h *AuthHandler) Me(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	user, err := h.svc.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toUserResponse(user))
}
