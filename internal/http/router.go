package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	appmiddleware "github.com/alpardfm/moneypath-api/internal/http/middleware"
)

// NewRouter creates the HTTP router used by the API.
func NewRouter(log *slog.Logger, healthHandler http.Handler) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(appmiddleware.RequestLogger(log))

	router.Method(http.MethodGet, "/health", healthHandler)

	return router
}
