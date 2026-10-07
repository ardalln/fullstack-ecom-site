package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"shop-api/internal/auth"
	"shop-api/internal/config"
	"shop-api/internal/handler"
	"shop-api/internal/postgres"
	"shop-api/internal/router"
	"shop-api/internal/service"
	"shop-api/internal/sms"
	"shop-api/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("application terminated", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	_ = godotenv.Load() // optional .env file for local development

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- infrastructure ---
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// --- wiring: repositories -> services -> handlers ---
	userRepo := postgres.NewUserRepository(pool)
	otpRepo := postgres.NewOTPRepository(pool)
	addressRepo := postgres.NewAddressRepository(pool)
	productRepo := postgres.NewProductRepository(pool)
	blogRepo := postgres.NewBlogRepository(pool)
	ticketRepo := postgres.NewTicketRepository(pool)
	commentRepo := postgres.NewCommentRepository(pool)
	orderRepo := postgres.NewOrderRepository(pool)
	bannerRepo := postgres.NewBannerRepository(pool)
	shippingRepo := postgres.NewShippingRepository(pool)
	couponRepo := postgres.NewCouponRepository(pool)
	paymentRepo := postgres.NewPaymentRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	brandRepo := postgres.NewBrandRepository(pool)
	dashboardRepo := postgres.NewDashboardRepository(pool)
	siteContentRepo := postgres.NewSiteContentRepository(pool)
	analyticsRepo := postgres.NewAnalyticsRepository(pool)
	txManager := postgres.NewTxManager(pool)

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)
	smsSender := sms.NewConsoleSender()

	authService := service.NewAuthService(userRepo, otpRepo, smsSender, tokens, service.OTPConfig{
		TTL:            cfg.OTPTTL,
		ResendCooldown: cfg.OTPResendCooldown,
		MaxAttempts:    cfg.OTPMaxAttempts,
	})
	addressService := service.NewAddressService(addressRepo)
	productService := service.NewProductService(productRepo, categoryRepo, brandRepo)
	blogService := service.NewBlogService(blogRepo)
	ticketService := service.NewTicketService(ticketRepo, userRepo, txManager)
	commentService := service.NewCommentService(commentRepo, productRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	brandService := service.NewBrandService(brandRepo)
	adminService := service.NewAdminService(userRepo, orderRepo, dashboardRepo, txManager)
	orderService := service.NewOrderService(orderRepo, addressRepo, txManager)
	commerceService := service.NewCommerceService(bannerRepo, shippingRepo, couponRepo)
	siteContentService := service.NewSiteContentService(siteContentRepo)
	paymentService := service.NewPaymentService(orderRepo, paymentRepo, txManager)

	if cfg.AdminPhone != "" {
		if err := authService.EnsureAdmin(ctx, cfg.AdminFirstName, cfg.AdminLastName, cfg.AdminPhone); err != nil {
			return fmt.Errorf("ensure admin user: %w", err)
		}
	}

	frontend, err := web.Static()
	if err != nil {
		return fmt.Errorf("load frontend files: %w", err)
	}

	e := router.New(router.Deps{
		Logger:      logger,
		Tokens:      tokens,
		Auth:        handler.NewAuthHandler(authService),
		Products:    handler.NewProductHandler(productService),
		Orders:      handler.NewOrderHandler(orderService),
		Addresses:   handler.NewAddressHandler(addressService),
		Payments:    handler.NewPaymentHandler(paymentService),
		Categories:  handler.NewCategoryHandler(categoryService),
		Brands:      handler.NewBrandHandler(brandService),
		Uploads:     handler.NewUploadHandler("uploads"),
		Comments:    handler.NewCommentHandler(commentService),
		Blog:        handler.NewBlogHandler(blogService),
		Tickets:     handler.NewTicketHandler(ticketService),
		Commerce:    handler.NewCommerceHandler(commerceService, orderRepo),
		SiteContent: handler.NewSiteContentHandler(siteContentService),
		Analytics:   handler.NewAnalyticsHandler(analyticsRepo),
		Admin:       handler.NewAdminHandler(adminService),
		Users:       userRepo,
		Static:      frontend,
	})
	e.Server.ReadHeaderTimeout = 5 * time.Second
	e.Server.ReadTimeout = 10 * time.Second
	e.Server.WriteTimeout = 15 * time.Second
	e.Server.IdleTimeout = 60 * time.Second

	// --- run + graceful shutdown ---
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", slog.String("port", cfg.AppPort))
		serverErr <- e.Start(":" + cfg.AppPort)
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
