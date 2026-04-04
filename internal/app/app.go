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
	"github.com/alpardfm/moneypath-api/internal/module/auth"
	"github.com/alpardfm/moneypath-api/internal/module/profile"
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

	healthHandler := handler.NewHealthHandler(db)
	authHandler := auth.NewHandler(authService)
	profileHandler := profile.NewHandler(profileService)
	authMiddleware := appmiddleware.NewAuthMiddleware(tokenManager)
	router := apihttp.NewRouter(log, healthHandler, authHandler, profileHandler, authMiddleware)

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
