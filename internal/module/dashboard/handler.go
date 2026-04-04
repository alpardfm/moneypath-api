package dashboard

import (
	"net/http"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves dashboard endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a dashboard handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get returns the derived dashboard overview.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	overview, err := h.service.GetOverview(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	response.Success(w, http.StatusOK, overview, nil)
}
