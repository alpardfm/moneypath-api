package mutation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/params"
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
		response.Unauthorized(w)
		return
	}
	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.InvalidJSON(w)
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
		response.Unauthorized(w)
		return
	}
	options, err := parseListOptions(r)
	if err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	result, err := h.service.List(r.Context(), userID, options)
	if err != nil {
		h.writeError(w, err)
		return
	}
	data := make([]map[string]any, 0, len(result.Items))
	for i := range result.Items {
		data = append(data, mutationResponse(&result.Items[i]))
	}
	response.Success(w, http.StatusOK, data, response.NewPaginationMeta(
		options.Page,
		options.PageSize,
		result.TotalItems,
	))
}

// GetByID returns a mutation detail.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
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
		response.Unauthorized(w)
		return
	}
	var input UpsertInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.InvalidJSON(w)
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
		response.Unauthorized(w)
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
		response.ValidationError(w, "wallet_id, type, amount, description, and happened_at are required")
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
		response.InternalError(w)
	}
}

func parseListOptions(r *http.Request) (ListOptions, error) {
	pagination := params.ParsePagination(r)
	options := ListOptions{
		Page:          pagination.Page,
		PageSize:      pagination.PageSize,
		Type:          strings.TrimSpace(r.URL.Query().Get("type")),
		WalletID:      strings.TrimSpace(r.URL.Query().Get("wallet_id")),
		DebtID:        strings.TrimSpace(r.URL.Query().Get("debt_id")),
		SortBy:        strings.TrimSpace(r.URL.Query().Get("sort_by")),
		SortDirection: strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort_direction"))),
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("related_to_debt")); raw != "" {
		switch strings.ToLower(raw) {
		case "true":
			value := true
			options.RelatedToDebt = &value
		case "false":
			value := false
			options.RelatedToDebt = &value
		default:
			return ListOptions{}, ErrMutationValidation
		}
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("from")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ListOptions{}, ErrMutationValidation
		}
		options.From = value.Format(time.RFC3339)
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("to")); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ListOptions{}, ErrMutationValidation
		}
		options.To = value.Format(time.RFC3339)
	}

	return options, nil
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
