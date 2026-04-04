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

type debtFixture struct {
	id        string
	userID    string
	principal string
	remaining string
	active    bool
}

type integrationRepo struct {
	wallets   map[string]*walletFixture
	debts     map[string]*debtFixture
	mutations map[string]Mutation
}

func (r *integrationRepo) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	wallet, ok := r.wallets[input.WalletID]
	if !ok || wallet.userID != userID || !wallet.active {
		return nil, ErrMutationWalletNotFound
	}

	var debtID *string
	debtAction := "none"
	switch {
	case !input.RelatedToDebt:
	case input.Type == "keluar":
		debt, ok := r.debts[*input.DebtID]
		if !ok || debt.userID != userID {
			return nil, ErrMutationDebtNotFound
		}
		if debt.remaining == "0.00" || debt.remaining == "0" {
			return nil, ErrDebtStateChanged
		}
		debt.remaining = "80.00"
		debtID = input.DebtID
		debtAction = "payment"
	case input.DebtID != nil:
		debt, ok := r.debts[*input.DebtID]
		if !ok || debt.userID != userID {
			return nil, ErrMutationDebtNotFound
		}
		debt.remaining = "170.00"
		debtID = input.DebtID
		debtAction = "borrow_existing"
	case input.NewDebt != nil:
		id := "debt-new"
		r.debts[id] = &debtFixture{
			id:        id,
			userID:    userID,
			principal: input.NewDebt.Principal,
			remaining: input.NewDebt.Principal,
			active:    true,
		}
		debtID = &id
		debtAction = "borrow_new"
	}

	switch input.Type {
	case "masuk":
		wallet.balance = "100.00"
	case "keluar":
		if wallet.balance == "0.00" || wallet.balance == "0" || input.Amount == "150.00" {
			return nil, ErrInsufficientWalletBalance
		}
		wallet.balance = "80.00"
	}

	item := Mutation{
		ID:            "mutation-1",
		UserID:        userID,
		WalletID:      input.WalletID,
		DebtID:        debtID,
		DebtAction:    debtAction,
		Type:          input.Type,
		Amount:        input.Amount,
		Description:   input.Description,
		RelatedToDebt: input.RelatedToDebt,
		HappenedAt:    input.HappenedAt,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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
	if item.DebtAction == "borrow_new" {
		delete(r.debts, *item.DebtID)
	}
	if item.DebtAction == "payment" && item.DebtID != nil {
		r.debts[*item.DebtID].remaining = "100.00"
	}
	if item.DebtAction == "borrow_existing" && item.DebtID != nil {
		r.debts[*item.DebtID].remaining = "50.00"
	}
	wallet := r.wallets[input.WalletID]
	if wallet == nil {
		return nil, ErrMutationWalletNotFound
	}
	wallet.balance = "0.00"

	created, err := r.Create(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	created.ID = mutationID
	created.CreatedAt = item.CreatedAt
	r.mutations[mutationID] = *created
	return created, nil
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

type noopDashboardRoutes struct{}

func (noopDashboardRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopSummaryRoutes struct{}

func (noopSummaryRoutes) Get(http.ResponseWriter, *http.Request) {}

func TestMutationDebtFlow(t *testing.T) {
	repo := &integrationRepo{
		wallets: map[string]*walletFixture{
			"wallet-1": {id: "wallet-1", userID: "user-1", balance: "50.00", active: true},
		},
		debts: map[string]*debtFixture{
			"debt-1": {id: "debt-1", userID: "user-1", principal: "100.00", remaining: "100.00", active: true},
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
		noopDashboardRoutes{},
		noopSummaryRoutes{},
		middleware.NewAuthMiddleware(tokenManager),
	)

	payReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","debt_id":"debt-1","type":"keluar","amount":"20.00","description":"pay installment","related_to_debt":true,"happened_at":"2026-04-04T10:00:00Z"}`))
	payReq.Header.Set("Content-Type", "application/json")
	payReq.Header.Set("Authorization", "Bearer "+token)
	payRes := httptest.NewRecorder()
	router.ServeHTTP(payRes, payReq)
	if payRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from debt payment, got %d", payRes.Code)
	}

	borrowNewReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"masuk","amount":"100.00","description":"loan disbursement","related_to_debt":true,"new_debt":{"name":"Laptop","principal_amount":"120.00","tenor_value":12,"tenor_unit":"month","payment_amount":"10.00"},"happened_at":"2026-04-04T11:00:00Z"}`))
	borrowNewReq.Header.Set("Content-Type", "application/json")
	borrowNewReq.Header.Set("Authorization", "Bearer "+token)
	borrowNewRes := httptest.NewRecorder()
	router.ServeHTTP(borrowNewRes, borrowNewReq)
	if borrowNewRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create debt by mutation, got %d", borrowNewRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/mutations/mutation-1", bytes.NewBufferString(`{"wallet_id":"wallet-1","debt_id":"debt-1","type":"masuk","amount":"120.00","description":"switch to existing debt borrow","related_to_debt":true,"happened_at":"2026-04-04T12:00:00Z"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from edit debt-related mutation, got %d", updateRes.Code)
	}
}
