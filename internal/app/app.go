package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alpardfm/moneypath-api/internal/config"
	apihttp "github.com/alpardfm/moneypath-api/internal/http"
	"github.com/alpardfm/moneypath-api/internal/http/handler"
	appmiddleware "github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/analytics"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
	"github.com/alpardfm/moneypath-api/internal/module/category"
	"github.com/alpardfm/moneypath-api/internal/module/dashboard"
	"github.com/alpardfm/moneypath-api/internal/module/debt"
	"github.com/alpardfm/moneypath-api/internal/module/export"
	"github.com/alpardfm/moneypath-api/internal/module/healthscore"
	"github.com/alpardfm/moneypath-api/internal/module/leakage"
	"github.com/alpardfm/moneypath-api/internal/module/mutation"
	"github.com/alpardfm/moneypath-api/internal/module/notification"
	"github.com/alpardfm/moneypath-api/internal/module/profile"
	"github.com/alpardfm/moneypath-api/internal/module/recurring"
	"github.com/alpardfm/moneypath-api/internal/module/settings"
	"github.com/alpardfm/moneypath-api/internal/module/summary"
	"github.com/alpardfm/moneypath-api/internal/module/wallet"
	"github.com/alpardfm/moneypath-api/internal/platform/database"
	"github.com/alpardfm/moneypath-api/internal/platform/logger"
)

// App wires the application dependencies.
type App struct {
	config *config.Config
	server *http.Server
	db     *database.Postgres
}

// New creates a fully wired application.
func New(cfg *config.Config) (*App, error) {
	log := logger.New(cfg.AppEnv)

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres connection: %w", err)
	}

	authRepo := auth.NewPostgresRepository(db.Pool())
	tokenManager := auth.NewTokenManager(cfg.JWTSecret)
	authService := auth.NewService(authRepo, tokenManager)
	profileService := profile.NewService(authRepo)
	settingsService := settings.NewService(authRepo)
	walletRepo := wallet.NewPostgresRepository(db.Pool())
	walletService := wallet.NewService(walletRepo)
	debtRepo := debt.NewPostgresRepository(db.Pool())
	debtService := debt.NewService(debtRepo)
	categoryRepo := category.NewPostgresRepository(db.Pool())
	categoryService := category.NewService(categoryRepo)
	mutationRepo := mutation.NewPostgresRepository(db.Pool())
	mutationService := mutation.NewService(mutationRepo)
	recurringRepo := recurring.NewPostgresRepository(db.Pool())
	recurringService := recurring.NewService(recurringRepo)
	analyticsRepo := analytics.NewPostgresRepository(db.Pool())
	analyticsService := analytics.NewService(analyticsRepo)
	exportRepo := export.NewPostgresRepository(db.Pool())
	exportService := export.NewService(exportRepo)
	healthScoreRepo := healthscore.NewPostgresRepository(db.Pool())
	healthScoreService := healthscore.NewService(healthScoreRepo)
	leakageRepo := leakage.NewPostgresRepository(db.Pool())
	leakageService := leakage.NewService(leakageRepo)
	notificationRepo := notification.NewPostgresRepository(db.Pool())
	notificationService := notification.NewService(notificationRepo)
	dashboardRepo := dashboard.NewPostgresRepository(db.Pool())
	dashboardService := dashboard.NewService(dashboardRepo)
	summaryRepo := summary.NewPostgresRepository(db.Pool())
	summaryService := summary.NewService(summaryRepo)

	healthHandler := handler.NewHealthHandler(db)
	authHandler := auth.NewHandler(authService)
	profileHandler := profile.NewHandler(profileService)
	settingsHandler := settings.NewHandler(settingsService)
	walletHandler := wallet.NewHandler(walletService)
	debtHandler := debt.NewHandler(debtService)
	categoryHandler := category.NewHandler(categoryService)
	mutationHandler := mutation.NewHandler(mutationService)
	recurringHandler := recurring.NewHandler(recurringService)
	analyticsHandler := analytics.NewHandler(analyticsService)
	exportHandler := export.NewHandler(exportService)
	healthScoreHandler := healthscore.NewHandler(healthScoreService)
	leakageHandler := leakage.NewHandler(leakageService)
	notificationHandler := notification.NewHandler(notificationService)
	dashboardHandler := dashboard.NewHandler(dashboardService)
	summaryHandler := summary.NewHandler(summaryService)
	authMiddleware := appmiddleware.NewAuthMiddleware(tokenManager)
	router := apihttp.NewRouter(log, healthHandler, authHandler, profileHandler, settingsHandler, walletHandler, debtHandler, categoryHandler, mutationHandler, recurringHandler, analyticsHandler, exportHandler, dashboardHandler, summaryHandler, healthScoreHandler, leakageHandler, notificationHandler, cfg.AllowedOrigins, authMiddleware)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		config: cfg,
		server: server,
		db:     db,
	}, nil
}

// Run starts the HTTP server and blocks until the context is canceled.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			errCh <- fmt.Errorf("shutdown server: %w", err)
			return
		}

		if a.db != nil {
			a.db.Close()
		}
	}()

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
