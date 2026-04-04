package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves auth endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates an auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register handles account creation.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.service.Register(r.Context(), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, map[string]any{
		"token": result.Token,
		"user":  userResponse(result.User),
	}, nil)
}

// Login handles account login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"token": result.Token,
		"user":  userResponse(result.User),
	}, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.Error(w, http.StatusBadRequest, "validation_error", "all fields are required")
	case errors.Is(err, ErrEmailAlreadyUsed):
		response.Error(w, http.StatusConflict, "email_already_used", err.Error())
	case errors.Is(err, ErrUsernameAlreadyUsed):
		response.Error(w, http.StatusConflict, "username_already_used", err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func userResponse(user *User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"full_name":  user.FullName,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
