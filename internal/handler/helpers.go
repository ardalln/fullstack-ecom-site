package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
	"shop-api/internal/middleware"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

// bindAndValidate decodes the JSON body into dst and validates its struct tags.
func bindAndValidate(c echo.Context, dst any) error {
	if err := c.Bind(dst); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	return c.Validate(dst)
}

// pagination reads ?page=&limit= with sane defaults and an upper bound.
func pagination(c echo.Context) (page, limit int) {
	page = queryInt(c, "page", defaultPage)
	limit = queryInt(c, "limit", defaultLimit)
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func queryInt(c echo.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.QueryParam(key))
	if err != nil {
		return fallback
	}
	return v
}

// idParam parses the ":id" path parameter.
func idParam(c echo.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: id must be a positive integer", domain.ErrInvalidInput)
	}
	return id, nil
}

// currentUserID returns the id of the authenticated user.
func currentUserID(c echo.Context) (int64, error) {
	id, ok := middleware.UserID(c)
	if !ok {
		return 0, domain.ErrUnauthorized
	}
	return id, nil
}
