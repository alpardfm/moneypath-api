package recurring

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/params"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves recurring endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a recurring handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles recurring rule creation.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.InvalidJSON(w)
		return
	}
	item, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, ruleResponse(item), nil)
}

// List returns active recurring rules.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	pagination := params.ParsePagination(r)
	result, err := h.service.List(r.Context(), userID, ListOptions{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Type:     r.URL.Query().Get("type"),
	})
	if err != nil {
		h.writeError(w, err)
		return
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, ruleResponse(&item))
	}
	response.Success(w, http.StatusOK, items, response.NewPaginationMeta(
		pagination.Page,
		pagination.PageSize,
		result.TotalItems,
	))
}

// GetByID returns one recurring rule.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	item, err := h.service.GetByID(r.Context(), userID, chi.URLParam(r, "ruleID"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, ruleResponse(item), nil)
}

// Update updates recurring rule metadata.
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
	item, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "ruleID"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, ruleResponse(item), nil)
}

// Inactivate hides a recurring rule.
func (h *Handler) Inactivate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	if err := h.service.Inactivate(r.Context(), userID, chi.URLParam(r, "ruleID")); err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"message": "recurring rule inactivated"}, nil)
}

// RunDue executes recurring rules due at the current time.
func (h *Handler) RunDue(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}
	result, err := h.service.RunDue(r.Context(), userID, time.Now().UTC())
	if err != nil {
		h.writeError(w, err)
		return
	}
	response.Success(w, http.StatusOK, result, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrRuleValidation):
		response.ValidationError(w, "wallet_id, type, amount, description, interval_unit, interval_step, and start_at are required")
	case errors.Is(err, ErrRuleNotFound):
		response.Error(w, http.StatusNotFound, "recurring_rule_not_found", err.Error())
	case errors.Is(err, ErrRuleWalletNotFound):
		response.Error(w, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, ErrRuleCategoryNotFound):
		response.Error(w, http.StatusNotFound, "category_not_found", err.Error())
	case errors.Is(err, ErrRuleCategoryMismatch):
		response.Error(w, http.StatusConflict, "category_type_mismatch", err.Error())
	default:
		response.InternalError(w)
	}
}

func ruleResponse(item *Rule) map[string]any {
	return map[string]any{
		"id":            item.ID,
		"user_id":       item.UserID,
		"wallet_id":     item.WalletID,
		"category_id":   item.CategoryID,
		"type":          item.Type,
		"amount":        item.Amount,
		"description":   item.Description,
		"interval_unit": item.IntervalUnit,
		"interval_step": item.IntervalStep,
		"start_at":      item.StartAt,
		"end_at":        item.EndAt,
		"next_run_at":   item.NextRunAt,
		"last_run_at":   item.LastRunAt,
		"is_active":     item.IsActive,
		"deleted_at":    item.DeletedAt,
		"created_at":    item.CreatedAt,
		"updated_at":    item.UpdatedAt,
	}
}
