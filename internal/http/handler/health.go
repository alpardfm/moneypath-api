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

	dbStatus := "up"
	if err := h.checker.Ping(ctx); err != nil {
		dbStatus = "down"
		response.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "degraded",
			"data": map[string]any{
				"service":  "moneypath-api",
				"database": dbStatus,
			},
			"error": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"data": map[string]any{
			"service":  "moneypath-api",
			"database": dbStatus,
		},
	})
}
