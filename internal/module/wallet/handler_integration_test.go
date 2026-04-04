package wallet

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	apihttp "github.com/alpardfm/moneypath-api/internal/http"
	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type integrationRepo struct {
	wallets map[string]Wallet
}

func (r *integrationRepo) Create(ctx context.Context, wallet *Wallet) error {
	if r.wallets == nil {
		r.wallets = map[string]Wallet{}
	}
	wallet.ID = "wallet-1"
	wallet.Balance = "0.00"
	wallet.IsActive = true
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = wallet.CreatedAt
	r.wallets[wallet.ID] = *wallet
	return nil
}

func (r *integrationRepo) ListActive(ctx context.Context, userID string) ([]Wallet, error) {
	var items []Wallet
	for _, wallet := range r.wallets {
		if wallet.UserID == userID && wallet.IsActive && wallet.DeletedAt == nil {
			items = append(items, wallet)
		}
	}
	return items, nil
}

func (r *integrationRepo) GetByID(ctx context.Context, userID, walletID string) (*Wallet, error) {
	wallet, ok := r.wallets[walletID]
	if !ok || wallet.UserID != userID {
		return nil, ErrWalletNotFound
	}
	return &wallet, nil
}

func (r *integrationRepo) UpdateName(ctx context.Context, userID, walletID, name string) (*Wallet, error) {
	wallet, ok := r.wallets[walletID]
	if !ok || wallet.UserID != userID {
		return nil, ErrWalletNotFound
	}
	wallet.Name = name
	wallet.UpdatedAt = time.Now()
	r.wallets[walletID] = wallet
	return &wallet, nil
}

func (r *integrationRepo) Inactivate(ctx context.Context, userID, walletID string) error {
	wallet, ok := r.wallets[walletID]
	if !ok || wallet.UserID != userID {
		return ErrWalletNotFound
	}
	if wallet.Balance != "0.00" && wallet.Balance != "0" {
		return ErrWalletBalanceNotZero
	}
	now := time.Now()
	wallet.IsActive = false
	wallet.DeletedAt = &now
	wallet.UpdatedAt = now
	r.wallets[walletID] = wallet
	return nil
}

type noopAuthRoutes struct{}

func (noopAuthRoutes) Register(http.ResponseWriter, *http.Request) {}
func (noopAuthRoutes) Login(http.ResponseWriter, *http.Request)    {}

type noopProfileRoutes struct{}

func (noopProfileRoutes) GetMe(http.ResponseWriter, *http.Request)          {}
func (noopProfileRoutes) UpdateMe(http.ResponseWriter, *http.Request)       {}
func (noopProfileRoutes) ChangePassword(http.ResponseWriter, *http.Request) {}

type noopDebtRoutes struct{}

func (noopDebtRoutes) Create(http.ResponseWriter, *http.Request)     {}
func (noopDebtRoutes) List(http.ResponseWriter, *http.Request)       {}
func (noopDebtRoutes) GetByID(http.ResponseWriter, *http.Request)    {}
func (noopDebtRoutes) Update(http.ResponseWriter, *http.Request)     {}
func (noopDebtRoutes) Inactivate(http.ResponseWriter, *http.Request) {}

type noopMutationRoutes struct{}

func (noopMutationRoutes) Create(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) List(http.ResponseWriter, *http.Request)    {}
func (noopMutationRoutes) GetByID(http.ResponseWriter, *http.Request) {}
func (noopMutationRoutes) Update(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) Delete(http.ResponseWriter, *http.Request)  {}

func TestWalletFlow(t *testing.T) {
	repo := &integrationRepo{}
	handler := NewHandler(NewService(repo))
	tokenManager := auth.NewTokenManager("secret")
	token, err := tokenManager.Generate("user-1")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := apihttp.NewRouter(
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		noopAuthRoutes{},
		noopProfileRoutes{},
		handler,
		noopDebtRoutes{},
		noopMutationRoutes{},
		middleware.NewAuthMiddleware(tokenManager),
	)

	createReq := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString(`{"name":"Cash"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create wallet, got %d", createRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/wallets", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from list wallets, got %d", listRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/wallets/wallet-1", bytes.NewBufferString(`{"name":"Main Cash"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from update wallet, got %d", updateRes.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/wallets/wallet-1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailRes := httptest.NewRecorder()
	router.ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from wallet detail, got %d", detailRes.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/wallets/wallet-1", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRes := httptest.NewRecorder()
	router.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from wallet inactivate, got %d", deleteRes.Code)
	}
}
