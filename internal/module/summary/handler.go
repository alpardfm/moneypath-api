package summary

import (
	"net/http"
	"time"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves summary endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a summary handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get returns the derived summary report.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	filter, err := parseFilter(r)
	if err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	report, err := h.service.GetReport(r.Context(), userID, filter)
	if err != nil {
		if err == ErrInvalidPeriod {
			response.ValidationError(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	response.Success(w, http.StatusOK, report, nil)
}

func parseFilter(r *http.Request) (Filter, error) {
	var filter Filter
	if from := r.URL.Query().Get("from"); from != "" {
		value, err := time.Parse("2006-01-02", from)
		if err != nil {
			return Filter{}, ErrInvalidPeriod
		}
		filter.From = &value
	}
	if to := r.URL.Query().Get("to"); to != "" {
		value, err := time.Parse("2006-01-02", to)
		if err != nil {
			return Filter{}, ErrInvalidPeriod
		}
		filter.To = &value
	}
	return filter, nil
}
