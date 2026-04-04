package profile

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

// Handler serves the profile endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a profile handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns the authenticated user profile.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, userResponse(user), nil)
}

// UpdateMe updates the authenticated user profile.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var input UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	user, err := h.service.UpdateMe(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, userResponse(user), nil)
}

// ChangePassword changes the authenticated user password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var input ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	if err := h.service.ChangePassword(r.Context(), userID, input); err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "password updated"}, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrValidation):
		response.Error(w, http.StatusBadRequest, "validation_error", "all fields are required")
	case errors.Is(err, auth.ErrUserNotFound):
		response.Error(w, http.StatusNotFound, "user_not_found", err.Error())
	case errors.Is(err, auth.ErrEmailAlreadyUsed):
		response.Error(w, http.StatusConflict, "email_already_used", err.Error())
	case errors.Is(err, auth.ErrUsernameAlreadyUsed):
		response.Error(w, http.StatusConflict, "username_already_used", err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func userResponse(user *auth.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"full_name":  user.FullName,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
