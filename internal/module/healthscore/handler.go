package healthscore

import (
	"net/http"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves financial health scoring endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a financial health scoring handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get returns the derived financial health score.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	report, err := h.service.GetReport(r.Context(), userID)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.Success(w, http.StatusOK, report, nil)
}
