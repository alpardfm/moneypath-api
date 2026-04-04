package mutation

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apihttp "github.com/alpardfm/moneypath-api/internal/http"
	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type walletFixture struct {
	id      string
	userID  string
	balance string
	active  bool
}

type integrationRepo struct {
	wallets   map[string]*walletFixture
	mutations map[string]Mutation
}

func (r *integrationRepo) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	wallet, ok := r.wallets[input.WalletID]
	if !ok || wallet.userID != userID || !wallet.active {
		return nil, ErrMutationWalletNotFound
	}
	if input.Type == "keluar" && wallet.balance == "0.00" {
		return nil, ErrInsufficientWalletBalance
	}
	if input.Type == "keluar" && wallet.balance == "100.00" && input.Amount == "150.00" {
		return nil, ErrInsufficientWalletBalance
	}
	if input.Type == "masuk" {
		wallet.balance = "100.00"
	}
	if input.Type == "keluar" && input.Amount == "20.00" {
		wallet.balance = "80.00"
	}

	item := Mutation{
		ID:          "mutation-1",
		UserID:      userID,
		WalletID:    input.WalletID,
		Type:        input.Type,
		Amount:      input.Amount,
		Description: input.Description,
		HappenedAt:  input.HappenedAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if r.mutations == nil {
		r.mutations = map[string]Mutation{}
	}
	r.mutations[item.ID] = item
	return &item, nil
}

func (r *integrationRepo) List(ctx context.Context, userID string) ([]Mutation, error) {
	var items []Mutation
	for _, item := range r.mutations {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *integrationRepo) GetByID(ctx context.Context, userID, mutationID string) (*Mutation, error) {
	item, ok := r.mutations[mutationID]
	if !ok || item.UserID != userID {
		return nil, ErrMutationNotFound
	}
	return &item, nil
}

func (r *integrationRepo) Update(ctx context.Context, userID, mutationID string, input UpsertInput) (*Mutation, error) {
	item, ok := r.mutations[mutationID]
	if !ok || item.UserID != userID {
		return nil, ErrMutationNotFound
	}
	wallet, ok := r.wallets[input.WalletID]
	if !ok || wallet.userID != userID || !wallet.active {
		return nil, ErrMutationWalletNotFound
	}
	if item.Type == "masuk" && input.Type == "keluar" && input.Amount == "150.00" {
		return nil, ErrInsufficientWalletBalance
	}
	item.WalletID = input.WalletID
	item.Type = input.Type
	item.Amount = input.Amount
	item.Description = input.Description
	item.HappenedAt = input.HappenedAt
	item.UpdatedAt = time.Now()
	r.mutations[mutationID] = item
	return &item, nil
}

func (r *integrationRepo) Delete(ctx context.Context, userID, mutationID string) error {
	if _, ok := r.mutations[mutationID]; !ok {
		return ErrMutationNotFound
	}
	return ErrMutationDeleteNotAllowed
}

type noopAuthRoutes struct{}

func (noopAuthRoutes) Register(http.ResponseWriter, *http.Request) {}
func (noopAuthRoutes) Login(http.ResponseWriter, *http.Request)    {}

type noopProfileRoutes struct{}

func (noopProfileRoutes) GetMe(http.ResponseWriter, *http.Request)          {}
func (noopProfileRoutes) UpdateMe(http.ResponseWriter, *http.Request)       {}
func (noopProfileRoutes) ChangePassword(http.ResponseWriter, *http.Request) {}

type noopWalletRoutes struct{}

func (noopWalletRoutes) Create(http.ResponseWriter, *http.Request)     {}
func (noopWalletRoutes) ListActive(http.ResponseWriter, *http.Request) {}
func (noopWalletRoutes) GetByID(http.ResponseWriter, *http.Request)    {}
func (noopWalletRoutes) Update(http.ResponseWriter, *http.Request)     {}
func (noopWalletRoutes) Inactivate(http.ResponseWriter, *http.Request) {}

type noopDebtRoutes struct{}

func (noopDebtRoutes) Create(http.ResponseWriter, *http.Request)     {}
func (noopDebtRoutes) List(http.ResponseWriter, *http.Request)       {}
func (noopDebtRoutes) GetByID(http.ResponseWriter, *http.Request)    {}
func (noopDebtRoutes) Update(http.ResponseWriter, *http.Request)     {}
func (noopDebtRoutes) Inactivate(http.ResponseWriter, *http.Request) {}

func TestMutationFlow(t *testing.T) {
	repo := &integrationRepo{
		wallets: map[string]*walletFixture{
			"wallet-1": {id: "wallet-1", userID: "user-1", balance: "0.00", active: true},
		},
		mutations: map[string]Mutation{},
	}
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
		noopWalletRoutes{},
		noopDebtRoutes{},
		handler,
		middleware.NewAuthMiddleware(tokenManager),
	)

	createIncomeReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"masuk","amount":"100.00","description":"salary","happened_at":"2026-04-04T10:00:00Z"}`))
	createIncomeReq.Header.Set("Content-Type", "application/json")
	createIncomeReq.Header.Set("Authorization", "Bearer "+token)
	createIncomeRes := httptest.NewRecorder()
	router.ServeHTTP(createIncomeRes, createIncomeReq)
	if createIncomeRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create income, got %d", createIncomeRes.Code)
	}

	createExpenseReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"keluar","amount":"150.00","description":"rent","happened_at":"2026-04-04T11:00:00Z"}`))
	createExpenseReq.Header.Set("Content-Type", "application/json")
	createExpenseReq.Header.Set("Authorization", "Bearer "+token)
	createExpenseRes := httptest.NewRecorder()
	router.ServeHTTP(createExpenseRes, createExpenseReq)
	if createExpenseRes.Code != http.StatusConflict {
		t.Fatalf("expected 409 from insufficient outgoing mutation, got %d", createExpenseRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/mutations/mutation-1", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"keluar","amount":"20.00","description":"groceries","happened_at":"2026-04-04T12:00:00Z"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from update mutation, got %d", updateRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/mutations", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from list mutations, got %d", listRes.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/mutations/mutation-1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailRes := httptest.NewRecorder()
	router.ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from mutation detail, got %d", detailRes.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/mutations/mutation-1", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRes := httptest.NewRecorder()
	router.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 from delete mutation, got %d", deleteRes.Code)
	}
}
