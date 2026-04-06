package notification

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

type integrationRepo struct{}

func (integrationRepo) ListDueRecurring(ctx context.Context, userID string, until time.Time) ([]RecurringAlert, error) {
	return []RecurringAlert{
		{RuleID: "rule-1", Amount: "100.00", Type: "keluar", NextRunAt: time.Now().Add(time.Hour)},
	}, nil
}

func (integrationRepo) ListActiveDebtAlerts(ctx context.Context, userID string, limit int) ([]DebtAlert, error) {
	return []DebtAlert{
		{DebtID: "debt-1", Name: "Laptop", RemainingAmount: "500.00", UpdatedAt: time.Now()},
	}, nil
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

func TestNotificationsEndpoint(t *testing.T) {
	handler := NewHandler(NewService(integrationRepo{}))
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
		noopRecurringRoutes{},
		noopAnalyticsRoutes{},
		noopExportRoutes{},
		noopDashboardRoutes{},
		noopSummaryRoutes{},
		noopHealthScoreRoutes{},
		noopLeakageRoutes{},
		handler,
		[]string{"http://localhost:5173"},
		middleware.NewAuthMiddleware(tokenManager),
	)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 from /notifications, got %d", res.Code)
	}
}
