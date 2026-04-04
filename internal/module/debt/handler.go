package debt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves debt endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a debt handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles debt creation from debt menu.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	item, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, debtResponse(item), nil)
}

// List returns debts for the authenticated user.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	data := make([]map[string]any, 0, len(items))
	for i := range items {
		data = append(data, debtResponse(&items[i]))
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetByID returns a debt detail.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	item, err := h.service.GetByID(r.Context(), userID, chi.URLParam(r, "debtID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, debtResponse(item), nil)
}

// Update updates debt metadata.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	item, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "debtID"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, debtResponse(item), nil)
}

// Inactivate soft deletes a paid debt.
func (h *Handler) Inactivate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	if err := h.service.Inactivate(r.Context(), userID, chi.URLParam(r, "debtID")); err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "debt inactivated"}, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrDebtValidation):
		response.Error(w, http.StatusBadRequest, "validation_error", "debt name and principal amount are required")
	case errors.Is(err, ErrDebtNotFound):
		response.Error(w, http.StatusNotFound, "debt_not_found", err.Error())
	case errors.Is(err, ErrDebtRemainingNotZero):
		response.Error(w, http.StatusConflict, "debt_remaining_not_zero", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func debtResponse(item *Debt) map[string]any {
	return map[string]any{
		"id":               item.ID,
		"user_id":          item.UserID,
		"name":             item.Name,
		"principal_amount": item.PrincipalAmount,
		"remaining_amount": item.RemainingAmount,
		"tenor_value":      item.TenorValue,
		"tenor_unit":       item.TenorUnit,
		"payment_amount":   item.PaymentAmount,
		"status":           item.Status,
		"is_active":        item.IsActive,
		"note":             item.Note,
		"deleted_at":       item.DeletedAt,
		"created_at":       item.CreatedAt,
		"updated_at":       item.UpdatedAt,
	}
}
