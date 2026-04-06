package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
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

type categoryFixture struct {
	id     string
	userID string
	typ    string
	active bool
}

type integrationRepo struct {
	categories map[string]*categoryFixture
	wallets    map[string]*walletFixture
	debts      map[string]*debtFixture
	mutations  map[string]Mutation
}

func (r *integrationRepo) Create(ctx context.Context, userID string, input UpsertInput) (*Mutation, error) {
	wallet, ok := r.wallets[input.WalletID]
	if !ok || wallet.userID != userID || !wallet.active {
		return nil, ErrMutationWalletNotFound
	}
	if input.CategoryID != nil {
		category, ok := r.categories[*input.CategoryID]
		if !ok || category.userID != userID || !category.active {
			return nil, ErrMutationCategoryNotFound
		}
		if category.typ != input.Type {
			return nil, ErrMutationCategoryMismatch
		}
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
		CategoryID:    input.CategoryID,
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

func (r *integrationRepo) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	var items []Mutation
	for _, item := range r.mutations {
		if item.UserID != userID {
			continue
		}
		if options.Type != "" && item.Type != options.Type {
			continue
		}
		if options.WalletID != "" && item.WalletID != options.WalletID {
			continue
		}
		if options.CategoryID != "" && (item.CategoryID == nil || *item.CategoryID != options.CategoryID) {
			continue
		}
		if options.DebtID != "" && (item.DebtID == nil || *item.DebtID != options.DebtID) {
			continue
		}
		if options.RelatedToDebt != nil && item.RelatedToDebt != *options.RelatedToDebt {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if options.SortBy == "amount" {
			if options.SortDirection == "asc" {
				return items[i].Amount < items[j].Amount
			}
			return items[i].Amount > items[j].Amount
		}
		if options.SortBy == "created_at" {
			if options.SortDirection == "asc" {
				return items[i].CreatedAt.Before(items[j].CreatedAt)
			}
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		if options.SortDirection == "asc" {
			return items[i].HappenedAt.Before(items[j].HappenedAt)
		}
		return items[i].HappenedAt.After(items[j].HappenedAt)
	})

	start := (options.Page - 1) * options.PageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + options.PageSize
	if end > len(items) {
		end = len(items)
	}

	return &ListResult{
		Items:      items[start:end],
		TotalItems: len(items),
	}, nil
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

type noopSettingsRoutes struct{}

func (noopSettingsRoutes) Get(http.ResponseWriter, *http.Request)    {}
func (noopSettingsRoutes) Update(http.ResponseWriter, *http.Request) {}

type noopWalletRoutes struct{}

func (noopWalletRoutes) Create(http.ResponseWriter, *http.Request)       {}
func (noopWalletRoutes) ListActive(http.ResponseWriter, *http.Request)   {}
func (noopWalletRoutes) ListArchived(http.ResponseWriter, *http.Request) {}
func (noopWalletRoutes) GetByID(http.ResponseWriter, *http.Request)      {}
func (noopWalletRoutes) Update(http.ResponseWriter, *http.Request)       {}
func (noopWalletRoutes) Inactivate(http.ResponseWriter, *http.Request)   {}

type noopDebtRoutes struct{}

func (noopDebtRoutes) Create(http.ResponseWriter, *http.Request)       {}
func (noopDebtRoutes) List(http.ResponseWriter, *http.Request)         {}
func (noopDebtRoutes) ListArchived(http.ResponseWriter, *http.Request) {}
func (noopDebtRoutes) GetByID(http.ResponseWriter, *http.Request)      {}
func (noopDebtRoutes) Update(http.ResponseWriter, *http.Request)       {}
func (noopDebtRoutes) Inactivate(http.ResponseWriter, *http.Request)   {}

type noopCategoryRoutes struct{}

func (noopCategoryRoutes) Create(http.ResponseWriter, *http.Request)     {}
func (noopCategoryRoutes) List(http.ResponseWriter, *http.Request)       {}
func (noopCategoryRoutes) GetByID(http.ResponseWriter, *http.Request)    {}
func (noopCategoryRoutes) Update(http.ResponseWriter, *http.Request)     {}
func (noopCategoryRoutes) Inactivate(http.ResponseWriter, *http.Request) {}

type noopRecurringRoutes struct{}

func (noopRecurringRoutes) Create(http.ResponseWriter, *http.Request)     {}
func (noopRecurringRoutes) List(http.ResponseWriter, *http.Request)       {}
func (noopRecurringRoutes) GetByID(http.ResponseWriter, *http.Request)    {}
func (noopRecurringRoutes) Update(http.ResponseWriter, *http.Request)     {}
func (noopRecurringRoutes) Inactivate(http.ResponseWriter, *http.Request) {}
func (noopRecurringRoutes) RunDue(http.ResponseWriter, *http.Request)     {}

type noopAnalyticsRoutes struct{}

func (noopAnalyticsRoutes) GetMonthly(http.ResponseWriter, *http.Request) {}

type noopExportRoutes struct{}

func (noopExportRoutes) ExportMutationsCSV(http.ResponseWriter, *http.Request) {}

type noopDashboardRoutes struct{}

func (noopDashboardRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopSummaryRoutes struct{}

func (noopSummaryRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopHealthScoreRoutes struct{}

func (noopHealthScoreRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopLeakageRoutes struct{}

func (noopLeakageRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopNotificationRoutes struct{}

func (noopNotificationRoutes) Get(http.ResponseWriter, *http.Request) {}

func TestMutationDebtFlow(t *testing.T) {
	repo := &integrationRepo{
		categories: map[string]*categoryFixture{
			"category-1": {id: "category-1", userID: "user-1", typ: "keluar", active: true},
			"category-2": {id: "category-2", userID: "user-1", typ: "masuk", active: true},
		},
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
		noopSettingsRoutes{},
		noopWalletRoutes{},
		noopDebtRoutes{},
		noopCategoryRoutes{},
		handler,
		noopRecurringRoutes{},
		noopAnalyticsRoutes{},
		noopExportRoutes{},
		noopDashboardRoutes{},
		noopSummaryRoutes{},
		noopHealthScoreRoutes{},
		noopLeakageRoutes{},
		noopNotificationRoutes{},
		[]string{"http://localhost:5173"},
		middleware.NewAuthMiddleware(tokenManager),
	)

	payReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","category_id":"category-1","debt_id":"debt-1","type":"keluar","amount":"20.00","description":"pay installment","related_to_debt":true,"happened_at":"2026-04-04T10:00:00Z"}`))
	payReq.Header.Set("Content-Type", "application/json")
	payReq.Header.Set("Authorization", "Bearer "+token)
	payRes := httptest.NewRecorder()
	router.ServeHTTP(payRes, payReq)
	if payRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from debt payment, got %d", payRes.Code)
	}

	borrowNewReq := httptest.NewRequest(http.MethodPost, "/mutations", bytes.NewBufferString(`{"wallet_id":"wallet-1","category_id":"category-2","type":"masuk","amount":"100.00","description":"loan disbursement","related_to_debt":true,"new_debt":{"name":"Laptop","principal_amount":"120.00","tenor_value":12,"tenor_unit":"month","payment_amount":"10.00"},"happened_at":"2026-04-04T11:00:00Z"}`))
	borrowNewReq.Header.Set("Content-Type", "application/json")
	borrowNewReq.Header.Set("Authorization", "Bearer "+token)
	borrowNewRes := httptest.NewRecorder()
	router.ServeHTTP(borrowNewRes, borrowNewReq)
	if borrowNewRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create debt by mutation, got %d", borrowNewRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/mutations/mutation-1", bytes.NewBufferString(`{"wallet_id":"wallet-1","category_id":"category-2","debt_id":"debt-1","type":"masuk","amount":"120.00","description":"switch to existing debt borrow","related_to_debt":true,"happened_at":"2026-04-04T12:00:00Z"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from edit debt-related mutation, got %d", updateRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/mutations?type=masuk&category_id=category-2&related_to_debt=true&sort_by=happened_at&sort_direction=desc&page=1&page_size=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from filtered mutation list, got %d", listRes.Code)
	}

	var listPayload struct {
		Data []struct {
			Type string `json:"type"`
		} `json:"data"`
		Meta struct {
			Page       int `json:"page"`
			PageSize   int `json:"page_size"`
			TotalItems int `json:"total_items"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(listRes.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if listPayload.Meta.Page != 1 || listPayload.Meta.PageSize != 10 || listPayload.Meta.TotalItems != 1 {
		t.Fatalf("unexpected list meta: %+v", listPayload.Meta)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].Type != "masuk" {
		t.Fatalf("unexpected filtered data: %+v", listPayload.Data)
	}
}
