package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/middleware"
)

// ErrorResponse is the JSON body of every error.
type ErrorResponse struct {
	Error      string            `json:"error"`
	Details    map[string]string `json:"details,omitempty"`
	RetryAfter *int              `json:"retry_after,omitempty"`
}

// NewErrorHandler returns Echo's central error handler. Handlers just return
// errors; this function decides the status code and the response body.
func NewErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status, body := mapError(err)

		if status >= http.StatusInternalServerError {
			// Details of unexpected errors are logged, never sent to the client.
			logger.ErrorContext(c.Request().Context(), "request failed",
				slog.String("method", c.Request().Method),
				slog.String("path", middleware.SafeLogPath(c.Request().URL.Path)),
				slog.String("error", err.Error()),
			)
		}
		if status == http.StatusUnauthorized {
			c.Response().Header().Set("WWW-Authenticate", "Bearer")
		}

		var respErr error
		if c.Request().Method == http.MethodHead {
			respErr = c.NoContent(status)
		} else {
			respErr = c.JSON(status, body)
		}
		if respErr != nil {
			logger.Error("failed to write error response", slog.String("error", respErr.Error()))
		}
	}
}

func mapError(err error) (int, ErrorResponse) {
	var (
		validationErrs validator.ValidationErrors
		httpErr        *echo.HTTPError
		fieldErr       *domain.FieldError
		cooldownErr    *domain.CooldownError
	)

	switch {
	case errors.As(err, &validationErrs):
		return http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "validation failed",
			Details: validationDetails(validationErrs),
		}
	case errors.As(err, &fieldErr):
		return http.StatusUnprocessableEntity, ErrorResponse{Error: "validation failed", Details: fieldErr.Fields}
	case errors.As(err, &cooldownErr):
		secs := cooldownErr.RetryAfterSeconds
		return http.StatusTooManyRequests, ErrorResponse{Error: err.Error(), RetryAfter: &secs}
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, ErrorResponse{Error: err.Error()}
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorResponse{Error: "missing or invalid access token"}
	case errors.Is(err, domain.ErrOTPNotRequested),
		errors.Is(err, domain.ErrOTPExpired),
		errors.Is(err, domain.ErrOTPInvalidCode),
		errors.Is(err, domain.ErrOTPTooManyAttempts):
		return http.StatusUnauthorized, ErrorResponse{Error: err.Error()}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, ErrorResponse{Error: domain.ErrForbidden.Error()}
	case errors.Is(err, domain.ErrAccountDisabled):
		return http.StatusForbidden, ErrorResponse{Error: domain.ErrAccountDisabled.Error()}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, ErrorResponse{Error: err.Error()}
	case errors.Is(err, domain.ErrPhoneTaken),
		errors.Is(err, domain.ErrUsernameTaken),
		errors.Is(err, domain.ErrEmailTaken),
		errors.Is(err, domain.ErrCategoryNameTaken),
		errors.Is(err, domain.ErrCategorySlugTaken),
		errors.Is(err, domain.ErrCategoryInUse),
		errors.Is(err, domain.ErrBrandNameTaken),
		errors.Is(err, domain.ErrBrandSlugTaken),
		errors.Is(err, domain.ErrBrandInUse),
		errors.Is(err, domain.ErrProductSlugTaken),
		errors.Is(err, domain.ErrProductVariantInUse),
		errors.Is(err, domain.ErrBlogSlugTaken),
		errors.Is(err, domain.ErrTicketClosed),
		errors.Is(err, domain.ErrLastAdmin),
		errors.Is(err, domain.ErrOrderStatusConflict),
		errors.Is(err, domain.ErrInsufficientStock),
		errors.Is(err, domain.ErrProductInUse),
		errors.Is(err, domain.ErrOrderNotPayable),
		errors.Is(err, domain.ErrPaymentAlreadyProcessed):
		return http.StatusConflict, ErrorResponse{Error: err.Error()}
	case errors.Is(err, domain.ErrCouponInvalid), errors.Is(err, domain.ErrCouponMinimum), errors.Is(err, domain.ErrShippingUnavailable):
		return http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()}
	case errors.As(err, &httpErr):
		return httpErr.Code, ErrorResponse{Error: fmt.Sprint(httpErr.Message)}
	default:
		return http.StatusInternalServerError, ErrorResponse{Error: "internal server error"}
	}
}
