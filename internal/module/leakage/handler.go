package leakage

import (
	"net/http"
	"strconv"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves leakage detection endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a leakage detection handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get returns the leakage detection report.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	days := 0
	if raw := r.URL.Query().Get("days"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			response.ValidationError(w, ErrInvalidDays.Error())
			return
		}
		days = value
	}

	report, err := h.service.GetReport(r.Context(), userID, days)
	if err != nil {
		if err == ErrInvalidDays {
			response.ValidationError(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.Success(w, http.StatusOK, report, nil)
}
