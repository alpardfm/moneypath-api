package recurring

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
	rules map[string]Rule
}

func (r *integrationRepo) Create(ctx context.Context, rule *Rule) error {
	if r.rules == nil {
		r.rules = map[string]Rule{}
	}
	rule.ID = "rule-1"
	rule.IsActive = true
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = rule.CreatedAt
	r.rules[rule.ID] = *rule
	return nil
}
func (r *integrationRepo) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	var items []Rule
	for _, item := range r.rules {
		if item.UserID == userID && item.IsActive && item.DeletedAt == nil {
			items = append(items, item)
		}
	}
	return &ListResult{Items: items, TotalItems: len(items)}, nil
}
func (r *integrationRepo) GetByID(ctx context.Context, userID, ruleID string) (*Rule, error) {
	item, ok := r.rules[ruleID]
	if !ok || item.UserID != userID {
		return nil, ErrRuleNotFound
	}
	return &item, nil
}
func (r *integrationRepo) Update(ctx context.Context, rule *Rule) (*Rule, error) {
	item, ok := r.rules[rule.ID]
	if !ok || item.UserID != rule.UserID {
		return nil, ErrRuleNotFound
	}
	rule.CreatedAt = item.CreatedAt
	rule.UpdatedAt = time.Now()
	rule.IsActive = true
	r.rules[rule.ID] = *rule
	return rule, nil
}
func (r *integrationRepo) Inactivate(ctx context.Context, userID, ruleID string) error {
	item, ok := r.rules[ruleID]
	if !ok || item.UserID != userID {
		return ErrRuleNotFound
	}
	now := time.Now()
	item.IsActive = false
	item.DeletedAt = &now
	r.rules[ruleID] = item
	return nil
}
func (r *integrationRepo) RunDue(ctx context.Context, userID string, now time.Time) (*RunDueResult, error) {
	return &RunDueResult{Processed: 1}, nil
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

type noopMutationRoutes struct{}

func (noopMutationRoutes) Create(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) List(http.ResponseWriter, *http.Request)    {}
func (noopMutationRoutes) GetByID(http.ResponseWriter, *http.Request) {}
func (noopMutationRoutes) Update(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) Delete(http.ResponseWriter, *http.Request)  {}

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

func TestRecurringFlow(t *testing.T) {
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
		noopSettingsRoutes{},
		noopWalletRoutes{},
		noopDebtRoutes{},
		noopCategoryRoutes{},
		noopMutationRoutes{},
		handler,
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

	createReq := httptest.NewRequest(http.MethodPost, "/recurring-rules", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"masuk","amount":"100.00","description":"salary","interval_unit":"monthly","interval_step":1,"start_at":"2026-04-01T10:00:00Z"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create recurring rule, got %d", createRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/recurring-rules", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from list recurring rules, got %d", listRes.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/recurring-rules/rule-1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailRes := httptest.NewRecorder()
	router.ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from recurring detail, got %d", detailRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/recurring-rules/rule-1", bytes.NewBufferString(`{"wallet_id":"wallet-1","type":"masuk","amount":"150.00","description":"salary updated","interval_unit":"monthly","interval_step":1,"start_at":"2026-04-01T10:00:00Z"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from update recurring rule, got %d", updateRes.Code)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/recurring-rules/run-due", nil)
	runReq.Header.Set("Authorization", "Bearer "+token)
	runRes := httptest.NewRecorder()
	router.ServeHTTP(runRes, runReq)
	if runRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from run due, got %d", runRes.Code)
	}
	var payload struct {
		Data RunDueResult `json:"data"`
	}
	if err := json.Unmarshal(runRes.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal run due response: %v", err)
	}
	if payload.Data.Processed != 1 {
		t.Fatalf("expected processed=1, got %+v", payload.Data)
	}
}
