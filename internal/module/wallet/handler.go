package wallet

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
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
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
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
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	wallets, err := h.service.ListActive(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(wallets))
	for _, wallet := range wallets {
		items = append(items, walletResponse(&wallet))
	}
	response.Success(w, http.StatusOK, items, nil)
}

// GetByID returns one wallet by id.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
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
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
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
		response.Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
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
		response.Error(w, http.StatusBadRequest, "validation_error", "wallet name is required")
	case errors.Is(err, ErrWalletNotFound):
		response.Error(w, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, ErrWalletNameAlreadyUsed):
		response.Error(w, http.StatusConflict, "wallet_name_already_used", err.Error())
	case errors.Is(err, ErrWalletBalanceNotZero):
		response.Error(w, http.StatusConflict, "wallet_balance_not_zero", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
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
