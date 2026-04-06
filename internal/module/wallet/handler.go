package wallet

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/params"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves wallet endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a wallet handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create handles wallet creation.
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

	wallet, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, walletResponse(wallet), nil)
}

// ListActive returns active wallets for the authenticated user.
func (h *Handler) ListActive(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	pagination := params.ParsePagination(r)
	result, err := h.service.ListActive(r.Context(), userID, ListOptions{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, wallet := range result.Items {
		items = append(items, walletResponse(&wallet))
	}
	response.Success(w, http.StatusOK, items, response.NewPaginationMeta(
		pagination.Page,
		pagination.PageSize,
		result.TotalItems,
	))
}

// ListArchived returns archived wallets for the authenticated user.
func (h *Handler) ListArchived(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	pagination := params.ParsePagination(r)
	result, err := h.service.ListArchived(r.Context(), userID, ListOptions{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, wallet := range result.Items {
		items = append(items, walletResponse(&wallet))
	}
	response.Success(w, http.StatusOK, items, response.NewPaginationMeta(
		pagination.Page,
		pagination.PageSize,
		result.TotalItems,
	))
}

// GetByID returns one wallet by id.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	wallet, err := h.service.GetByID(r.Context(), userID, chi.URLParam(r, "walletID"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, walletResponse(wallet), nil)
}

// Update updates wallet metadata.
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

	wallet, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "walletID"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, walletResponse(wallet), nil)
}

// Inactivate inactivates a wallet when the balance is zero.
func (h *Handler) Inactivate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	err := h.service.Inactivate(r.Context(), userID, chi.URLParam(r, "walletID"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "wallet inactivated"}, nil)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrWalletValidation):
		response.ValidationError(w, "wallet name is required")
	case errors.Is(err, ErrWalletNotFound):
		response.Error(w, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, ErrWalletNameAlreadyUsed):
		response.Error(w, http.StatusConflict, "wallet_name_already_used", err.Error())
	case errors.Is(err, ErrWalletBalanceNotZero):
		response.Error(w, http.StatusConflict, "wallet_balance_not_zero", err.Error())
	default:
		response.InternalError(w)
	}
}

func walletResponse(wallet *Wallet) map[string]any {
	return map[string]any{
		"id":         wallet.ID,
		"user_id":    wallet.UserID,
		"name":       wallet.Name,
		"balance":    wallet.Balance,
		"is_active":  wallet.IsActive,
		"deleted_at": wallet.DeletedAt,
		"created_at": wallet.CreatedAt,
		"updated_at": wallet.UpdatedAt,
	}
}
