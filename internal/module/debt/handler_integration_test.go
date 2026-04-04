package debt

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apihttp "github.com/alpardfm/moneypath-api/internal/http"
	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type integrationRepo struct {
	debts map[string]Debt
}

func (r *integrationRepo) Create(ctx context.Context, debt *Debt) error {
	if r.debts == nil {
		r.debts = map[string]Debt{}
	}
	debt.ID = "debt-1"
	debt.IsActive = true
	debt.CreatedAt = time.Now()
	debt.UpdatedAt = debt.CreatedAt
	r.debts[debt.ID] = *debt
	return nil
}
func (r *integrationRepo) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	var out []Debt
	for _, debt := range r.debts {
		if debt.UserID == userID {
			out = append(out, debt)
		}
	}
	start := (options.Page - 1) * options.PageSize
	if start > len(out) {
		start = len(out)
	}
	end := start + options.PageSize
	if end > len(out) {
		end = len(out)
	}
	return &ListResult{
		Items:      out[start:end],
		TotalItems: len(out),
	}, nil
}
func (r *integrationRepo) GetByID(ctx context.Context, userID, debtID string) (*Debt, error) {
	debt, ok := r.debts[debtID]
	if !ok || debt.UserID != userID {
		return nil, ErrDebtNotFound
	}
	return &debt, nil
}
func (r *integrationRepo) Update(ctx context.Context, debt *Debt) (*Debt, error) {
	current, ok := r.debts[debt.ID]
	if !ok || current.UserID != debt.UserID {
		return nil, ErrDebtNotFound
	}
	current.Name = debt.Name
	current.TenorValue = debt.TenorValue
	current.TenorUnit = debt.TenorUnit
	current.PaymentAmount = debt.PaymentAmount
	current.Note = debt.Note
	current.UpdatedAt = time.Now()
	current.Status = deriveStatus(current.RemainingAmount, current.IsActive)
	r.debts[debt.ID] = current
	return &current, nil
}
func (r *integrationRepo) Inactivate(ctx context.Context, userID, debtID string) error {
	current, ok := r.debts[debtID]
	if !ok || current.UserID != userID {
		return ErrDebtNotFound
	}
	if current.RemainingAmount != "0.00" && current.RemainingAmount != "0" {
		return ErrDebtRemainingNotZero
	}
	now := time.Now()
	current.IsActive = false
	current.DeletedAt = &now
	current.Status = "inactive"
	current.UpdatedAt = now
	r.debts[debtID] = current
	return nil
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

type noopMutationRoutes struct{}

func (noopMutationRoutes) Create(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) List(http.ResponseWriter, *http.Request)    {}
func (noopMutationRoutes) GetByID(http.ResponseWriter, *http.Request) {}
func (noopMutationRoutes) Update(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) Delete(http.ResponseWriter, *http.Request)  {}

type noopDashboardRoutes struct{}

func (noopDashboardRoutes) Get(http.ResponseWriter, *http.Request) {}

type noopSummaryRoutes struct{}

func (noopSummaryRoutes) Get(http.ResponseWriter, *http.Request) {}

func TestDebtFlow(t *testing.T) {
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
		noopWalletRoutes{},
		handler,
		noopMutationRoutes{},
		noopDashboardRoutes{},
		noopSummaryRoutes{},
		middleware.NewAuthMiddleware(tokenManager),
	)

	createReq := httptest.NewRequest(http.MethodPost, "/debts", bytes.NewBufferString(`{"name":"Laptop","principal_amount":"1200.00","tenor_value":12,"tenor_unit":"month","payment_amount":"100.00"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create debt, got %d", createRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/debts", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from list debts, got %d", listRes.Code)
	}
	var listPayload struct {
		Meta struct {
			Page       int `json:"page"`
			PageSize   int `json:"page_size"`
			TotalItems int `json:"total_items"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(listRes.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if listPayload.Meta.Page != 1 || listPayload.Meta.PageSize != 20 || listPayload.Meta.TotalItems != 1 {
		t.Fatalf("unexpected pagination meta: %+v", listPayload.Meta)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/debts/debt-1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailRes := httptest.NewRecorder()
	router.ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from debt detail, got %d", detailRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/debts/debt-1", bytes.NewBufferString(`{"name":"Laptop Cicilan","tenor_value":10,"tenor_unit":"month","payment_amount":"120.00"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from update debt, got %d", updateRes.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/debts/debt-1", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRes := httptest.NewRecorder()
	router.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusConflict {
		t.Fatalf("expected 409 from delete unpaid debt, got %d", deleteRes.Code)
	}
}
