package settings

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

// Handler serves settings endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a settings handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Get returns the current user settings.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	user, err := h.service.Get(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, settingsResponse(user), nil)
}

// Update updates the current user settings.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.InvalidJSON(w)
		return
	}
	user, err := h.service.Update(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, settingsResponse(user), nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrValidation):
		response.ValidationError(w, "preferred_currency, timezone, date_format, and week_start_day are required")
	case errors.Is(err, auth.ErrUserNotFound):
		response.Error(w, http.StatusNotFound, "user_not_found", err.Error())
	default:
		response.InternalError(w)
	}
}

func settingsResponse(user *auth.User) map[string]any {
	return map[string]any{
		"preferred_currency": user.PreferredCurrency,
		"timezone":           user.Timezone,
		"date_format":        user.DateFormat,
		"week_start_day":     user.WeekStartDay,
		"updated_at":         user.UpdatedAt,
	}
}
