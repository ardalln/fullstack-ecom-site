package router

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"shop-api/internal/auth"
	"shop-api/internal/domain"
	"shop-api/internal/handler"
	"shop-api/internal/middleware"
)

// Deps are everything the router needs to wire up the routes.
type Deps struct {
	Logger      *slog.Logger
	Tokens      *auth.TokenManager
	Auth        *handler.AuthHandler
	Products    *handler.ProductHandler
	Orders      *handler.OrderHandler
	Addresses   *handler.AddressHandler
	Payments    *handler.PaymentHandler
	Categories  *handler.CategoryHandler
	Admin       *handler.AdminHandler
	Brands      *handler.BrandHandler
	Uploads     *handler.UploadHandler
	Comments    *handler.CommentHandler
	Blog        *handler.BlogHandler
	Tickets     *handler.TicketHandler
	Commerce    *handler.CommerceHandler
	SiteContent *handler.SiteContentHandler
	Analytics   *handler.AnalyticsHandler
	Users       domain.UserRepository

	// Static holds the frontend files. When nil, only the API is served.
	Static fs.FS
}

func New(d Deps) *echo.Echo {
	e := echo.New()
	// Do not trust client-supplied X-Forwarded-For headers for rate limiting.
	// Configure a trusted proxy extractor explicitly if this runs behind one.
	e.IPExtractor = echo.ExtractIPDirect()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = handler.NewValidator()
	e.HTTPErrorHandler = handler.NewErrorHandler(d.Logger)

	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.BodyLimit("8M"))
	e.Use(securityHeaders)
	e.Use(requestLogger(d.Logger))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	authRequired := middleware.Auth(d.Tokens, d.Users)
	adminOnly := middleware.RequireRole(domain.RoleAdmin)

	api := e.Group("/api/v1")

	// Public
	otpRequestLimiter := echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{Rate: 1.0 / 30, Burst: 2, ExpiresIn: 10 * time.Minute})
	otpVerifyLimiter := echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{Rate: 1.0 / 10, Burst: 5, ExpiresIn: 10 * time.Minute})
	api.POST("/auth/otp/request", d.Auth.RequestOTP, echomw.RateLimiter(otpRequestLimiter))
	api.POST("/auth/otp/verify", d.Auth.VerifyOTP, echomw.RateLimiter(otpVerifyLimiter))
	api.GET("/products", d.Products.List)
	api.GET("/products/slug/:slug", d.Products.GetBySlug)
	api.GET("/products/:id/comments", d.Comments.ListApproved)
	api.GET("/products/:id", d.Products.Get)
	api.GET("/categories", d.Categories.List)
	api.GET("/categories/slug/:slug", d.Categories.GetBySlug)
	api.GET("/brands", d.Brands.List)
	api.GET("/brands/slug/:slug", d.Brands.GetBySlug)
	api.GET("/site-content", d.SiteContent.Get)
	api.GET("/blog/posts", d.Blog.ListPublished)
	api.GET("/blog/posts/:slug", d.Blog.GetPublished)
	api.GET("/banners", d.Commerce.ListBanners)
	api.GET("/shipping-methods", d.Commerce.ListShipping)
	api.GET("/tracking/:code", d.Commerce.Track)
	analyticsLimiter := echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{Rate: 1.0 / 2, Burst: 6, ExpiresIn: 10 * time.Minute})
	api.POST("/analytics/heartbeat", d.Analytics.Heartbeat, echomw.RateLimiter(analyticsLimiter))
	api.POST("/analytics/view", d.Analytics.RecordView, echomw.RateLimiter(analyticsLimiter))
	// The payment gateway routes are public on purpose: the random authority
	// token in the URL is the capability, just like a real gateway's
	// checkout link needs no separate merchant login.
	api.GET("/payments/:authority", d.Payments.Gateway)
	api.POST("/payments/:authority/confirm", d.Payments.Confirm)

	// Authenticated users
	api.GET("/me", d.Auth.Me, authRequired)
	api.POST("/coupons/validate", d.Commerce.ValidateCoupon, authRequired)
	api.GET("/addresses", d.Addresses.List, authRequired)
	api.POST("/addresses", d.Addresses.Create, authRequired)
	api.PUT("/addresses/:id", d.Addresses.Update, authRequired)
	api.DELETE("/addresses/:id", d.Addresses.Delete, authRequired)
	api.PATCH("/addresses/:id/default", d.Addresses.SetDefault, authRequired)
	api.POST("/orders", d.Orders.Create, authRequired)
	api.POST("/products/:id/comments", d.Comments.Create, authRequired)
	api.GET("/orders", d.Orders.List, authRequired)
	api.GET("/orders/:id", d.Orders.Get, authRequired)
	api.POST("/orders/:id/pay", d.Payments.Request, authRequired)
	api.GET("/tickets", d.Tickets.ListMine, authRequired)
	api.POST("/tickets", d.Tickets.Create, authRequired)
	api.GET("/tickets/:id", d.Tickets.GetMine, authRequired)
	api.POST("/tickets/:id/messages", d.Tickets.ReplyMine, authRequired)

	// Admins only
	api.POST("/products", d.Products.Create, authRequired, adminOnly)
	api.PUT("/products/:id", d.Products.Update, authRequired, adminOnly)
	api.DELETE("/products/:id", d.Products.Delete, authRequired, adminOnly)
	api.GET("/admin/products", d.Products.ListAdmin, authRequired, adminOnly)
	api.PATCH("/admin/products/:id/visibility", d.Products.SetVisibility, authRequired, adminOnly)
	api.GET("/admin/blog/posts", d.Blog.ListAdmin, authRequired, adminOnly)
	api.GET("/admin/blog/posts/:id", d.Blog.GetAdmin, authRequired, adminOnly)
	api.POST("/admin/blog/posts", d.Blog.Create, authRequired, adminOnly)
	api.PUT("/admin/blog/posts/:id", d.Blog.Update, authRequired, adminOnly)
	api.DELETE("/admin/blog/posts/:id", d.Blog.Delete, authRequired, adminOnly)
	api.GET("/admin/tickets", d.Tickets.ListAdmin, authRequired, adminOnly)
	api.GET("/admin/tickets/:id", d.Tickets.GetAdmin, authRequired, adminOnly)
	api.POST("/admin/tickets/:id/messages", d.Tickets.ReplyAdmin, authRequired, adminOnly)
	api.PATCH("/admin/tickets/:id/status", d.Tickets.SetStatus, authRequired, adminOnly)
	api.POST("/admin/uploads", d.Uploads.Image, authRequired, adminOnly)
	api.GET("/admin/overview", d.Admin.Overview, authRequired, adminOnly)
	api.GET("/admin/analytics", d.Analytics.Overview, authRequired, adminOnly)
	api.PUT("/admin/site-content", d.SiteContent.Update, authRequired, adminOnly)
	api.GET("/admin/banners", d.Commerce.ListAdminBanners, authRequired, adminOnly)
	api.POST("/admin/banners", d.Commerce.SaveBanner, authRequired, adminOnly)
	api.PUT("/admin/banners/:id", d.Commerce.SaveBanner, authRequired, adminOnly)
	api.DELETE("/admin/banners/:id", d.Commerce.DeleteBanner, authRequired, adminOnly)
	api.GET("/admin/shipping-methods", d.Commerce.ListAdminShipping, authRequired, adminOnly)
	api.POST("/admin/shipping-methods", d.Commerce.SaveShipping, authRequired, adminOnly)
	api.PUT("/admin/shipping-methods/:id", d.Commerce.SaveShipping, authRequired, adminOnly)
	api.DELETE("/admin/shipping-methods/:id", d.Commerce.DeleteShipping, authRequired, adminOnly)
	api.GET("/admin/coupons", d.Commerce.ListCoupons, authRequired, adminOnly)
	api.POST("/admin/coupons", d.Commerce.SaveCoupon, authRequired, adminOnly)
	api.PUT("/admin/coupons/:id", d.Commerce.SaveCoupon, authRequired, adminOnly)
	api.DELETE("/admin/coupons/:id", d.Commerce.DeleteCoupon, authRequired, adminOnly)
	api.GET("/admin/users", d.Admin.ListUsers, authRequired, adminOnly)
	api.PATCH("/admin/users/:id/access", d.Admin.ChangeUserAccess, authRequired, adminOnly)
	api.GET("/admin/orders", d.Admin.ListOrders, authRequired, adminOnly)
	api.GET("/admin/comments", d.Comments.ListAll, authRequired, adminOnly)
	api.PATCH("/admin/comments/:id/approval", d.Comments.SetApproval, authRequired, adminOnly)
	api.PATCH("/admin/orders/:id/status", d.Admin.ChangeOrderStatus, authRequired, adminOnly)
	api.POST("/admin/categories", d.Categories.Create, authRequired, adminOnly)
	api.PUT("/admin/categories/:id", d.Categories.Update, authRequired, adminOnly)
	api.DELETE("/admin/categories/:id", d.Categories.Delete, authRequired, adminOnly)
	api.POST("/admin/brands", d.Brands.Create, authRequired, adminOnly)
	api.PUT("/admin/brands/:id", d.Brands.Update, authRequired, adminOnly)
	api.DELETE("/admin/brands/:id", d.Brands.Delete, authRequired, adminOnly)

	// Unknown API paths get a JSON 404 instead of falling through to the frontend.
	api.Any("/*", func(c echo.Context) error { return echo.ErrNotFound })

	if d.Static != nil {
		e.Static("/uploads", "uploads")
		e.GET("/blog", d.Blog.SEOIndex)
		e.GET("/blog/:slug", d.Blog.SEOPage)
		serveFrontend(e, d.Static)
	}

	return e
}

func securityHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		h := c.Response().Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; frame-ancestors 'none'; object-src 'none'; form-action 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; img-src 'self' https: data:; connect-src 'self'")
		return next(c)
	}
}

// serveFrontend serves the single-page frontend (index.html, app.js, styles.css)
// from the same origin as the API, so no CORS configuration is needed.
func serveFrontend(e *echo.Echo, files fs.FS) {
	fileServer := http.FileServer(http.FS(files))
	serve := func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-cache")
		filePath := strings.TrimPrefix(c.Request().URL.Path, "/")
		if filePath != "" {
			if _, err := fs.Stat(files, filePath); err != nil {
				if strings.Contains(c.Request().Header.Get(echo.HeaderAccept), "text/html") {
					index, readErr := fs.ReadFile(files, "index.html")
					if readErr != nil {
						return fmt.Errorf("read frontend entrypoint: %w", readErr)
					}
					return c.Blob(http.StatusOK, "text/html; charset=utf-8", index)
				}
				return echo.ErrNotFound
			}
		}
		fileServer.ServeHTTP(c.Response(), c.Request())
		return nil
	}
	e.GET("/", serve)
	e.GET("/*", serve)
}

func requestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		// Only log API calls, not requests for the frontend's static files.
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().URL.Path, "/api/")
		},
		LogMethod:    true,
		LogURI:       false,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogError:     true,
		HandleError:  true, // let the central error handler set the real status first
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			attrs := []slog.Attr{
				slog.String("request_id", v.RequestID),
				slog.String("method", v.Method),
				slog.String("path", middleware.SafeLogPath(c.Request().URL.Path)),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			}
			level := slog.LevelInfo
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
				if v.Status >= http.StatusInternalServerError {
					level = slog.LevelError
				}
			}
			logger.LogAttrs(c.Request().Context(), level, "http_request", attrs...)
			return nil
		},
	})
}
