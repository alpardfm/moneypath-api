package analytics

import (
	"net/http"
	"strconv"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves analytics endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates an analytics handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMonthly returns the monthly analytics report.
func (h *Handler) GetMonthly(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	months := 0
	if raw := r.URL.Query().Get("months"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			response.ValidationError(w, ErrInvalidMonths.Error())
			return
		}
		months = value
	}

	report, err := h.service.GetMonthlyReport(r.Context(), userID, months)
	if err != nil {
		if err == ErrInvalidMonths {
			response.ValidationError(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.Success(w, http.StatusOK, report, nil)
}
