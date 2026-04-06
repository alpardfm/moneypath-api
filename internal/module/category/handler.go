package category

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/params"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves category endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a category handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles category creation.
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

	response.Success(w, http.StatusCreated, categoryResponse(item), nil)
}

// List returns active categories for the authenticated user.
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
		items = append(items, categoryResponse(&item))
	}
	response.Success(w, http.StatusOK, items, response.NewPaginationMeta(
		pagination.Page,
		pagination.PageSize,
		result.TotalItems,
	))
}

// GetByID returns one category by id.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	item, err := h.service.GetByID(r.Context(), userID, chi.URLParam(r, "categoryID"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, categoryResponse(item), nil)
}

// Update updates category metadata.
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

	item, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "categoryID"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, categoryResponse(item), nil)
}

// Inactivate hides a category from active selections.
func (h *Handler) Inactivate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	if err := h.service.Inactivate(r.Context(), userID, chi.URLParam(r, "categoryID")); err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "category inactivated"}, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCategoryValidation):
		response.ValidationError(w, "category name and type are required, and type must be masuk or keluar")
	case errors.Is(err, ErrCategoryNotFound):
		response.Error(w, http.StatusNotFound, "category_not_found", err.Error())
	case errors.Is(err, ErrCategoryNameAlreadyUsed):
		response.Error(w, http.StatusConflict, "category_name_already_used", err.Error())
	default:
		response.InternalError(w)
	}
}

func categoryResponse(item *Category) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"user_id":    item.UserID,
		"name":       item.Name,
		"type":       item.Type,
		"is_active":  item.IsActive,
		"deleted_at": item.DeletedAt,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	}
}
