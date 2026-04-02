package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// HealthChecker captures the dependency required by the health endpoint.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// HealthHandler serves the health check endpoint.
type HealthHandler struct {
	checker HealthChecker
}

// NewHealthHandler creates a health handler instance.
func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP returns API and database availability information.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.checker.Ping(ctx); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "database_unavailable", err.Error())
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"service":  "moneypath-api",
		"database": "up",
	}, nil)
}
