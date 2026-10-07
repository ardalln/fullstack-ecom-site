package middleware

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"shop-api/internal/auth"
	"shop-api/internal/domain"
)

const (
	ctxUserIDKey = "auth.user_id"
	ctxRoleKey   = "auth.role"
)

// Auth requires a valid "Authorization: Bearer <jwt>" header and stores the
// user id and role in the request context.
func Auth(tokens *auth.TokenManager, users domain.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			scheme, raw, ok := strings.Cut(c.Request().Header.Get(echo.HeaderAuthorization), " ")
			raw = strings.TrimSpace(raw)
			if !ok || !strings.EqualFold(scheme, "Bearer") || raw == "" {
				return domain.ErrUnauthorized
			}

			claims, err := tokens.Parse(raw)
			if err != nil {
				return domain.ErrUnauthorized
			}
			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil || userID <= 0 {
				return domain.ErrUnauthorized
			}

			role := claims.Role
			if users != nil {
				user, err := users.GetByID(c.Request().Context(), userID)
				if err != nil || !user.IsActive {
					return domain.ErrUnauthorized
				}
				// Make role changes and suspensions effective for existing JWTs.
				role = user.Role
			}
			c.Set(ctxUserIDKey, userID)
			c.Set(ctxRoleKey, role)
			return next(c)
		}
	}
}

// RequireRole allows the request only if the authenticated user has one of the
// given roles. It must run after Auth.
func RequireRole(roles ...domain.Role) echo.MiddlewareFunc {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(ctxRoleKey).(domain.Role)
			if !ok {
				return domain.ErrUnauthorized
			}
			if _, permitted := allowed[role]; !permitted {
				return domain.ErrForbidden
			}
			return next(c)
		}
	}
}

// UserID returns the authenticated user's id set by the Auth middleware.
func UserID(c echo.Context) (int64, bool) {
	id, ok := c.Get(ctxUserIDKey).(int64)
	return id, ok
}
