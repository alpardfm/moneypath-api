package mutation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves mutation endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a mutation handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles incoming and outgoing mutation creation.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	item, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, mutationResponse(item), nil)
}

// List returns mutation history.
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
		data = append(data, mutationResponse(&items[i]))
	}
	response.Success(w, http.StatusOK, data, nil)
}

// GetByID returns a mutation detail.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	item, err := h.service.GetByID(r.Context(), userID, chi.URLParam(r, "mutationID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, mutationResponse(item), nil)
}

// Update edits an existing mutation.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}
	item, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "mutationID"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, mutationResponse(item), nil)
}

// Delete explicitly rejects mutation deletion.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "mutationID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMutationValidation):
		response.Error(w, http.StatusBadRequest, "validation_error", "wallet_id, type, amount, description, and happened_at are required")
	case errors.Is(err, ErrInvalidDebtRelation):
		response.Error(w, http.StatusBadRequest, "invalid_debt_relation", err.Error())
	case errors.Is(err, ErrMutationNotFound):
		response.Error(w, http.StatusNotFound, "mutation_not_found", err.Error())
	case errors.Is(err, ErrMutationWalletNotFound):
		response.Error(w, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, ErrMutationDebtNotFound):
		response.Error(w, http.StatusNotFound, "debt_not_found", err.Error())
	case errors.Is(err, ErrInsufficientWalletBalance):
		response.Error(w, http.StatusConflict, "insufficient_wallet_balance", err.Error())
	case errors.Is(err, ErrDebtStateChanged):
		response.Error(w, http.StatusConflict, "debt_state_changed", err.Error())
	case errors.Is(err, ErrMutationDeleteNotAllowed):
		response.Error(w, http.StatusMethodNotAllowed, "mutation_delete_not_allowed", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func mutationResponse(item *Mutation) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"user_id":         item.UserID,
		"wallet_id":       item.WalletID,
		"debt_id":         item.DebtID,
		"debt_action":     item.DebtAction,
		"type":            item.Type,
		"amount":          item.Amount,
		"description":     item.Description,
		"related_to_debt": item.RelatedToDebt,
		"happened_at":     item.HappenedAt,
		"created_at":      item.CreatedAt,
		"updated_at":      item.UpdatedAt,
	}
}
