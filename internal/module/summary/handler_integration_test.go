package summary

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	apihttp "github.com/alpardfm/moneypath-api/internal/http"
	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type integrationRepo struct{}

func (integrationRepo) GetReport(ctx context.Context, userID string, filter Filter) (*Report, error) {
	return &Report{
		TotalAssets:   "100.00",
		TotalDebts:    "20.00",
		TotalIncoming: "200.00",
		TotalOutgoing: "100.00",
		NetFlow:       "100.00",
		Wallets:       []WalletBalance{{WalletID: "wallet-1", Name: "Cash", Balance: "100.00"}},
		From:          filter.From,
		To:            filter.To,
	}, nil
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

type noopMutationRoutes struct{}

func (noopMutationRoutes) Create(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) List(http.ResponseWriter, *http.Request)    {}
func (noopMutationRoutes) GetByID(http.ResponseWriter, *http.Request) {}
func (noopMutationRoutes) Update(http.ResponseWriter, *http.Request)  {}
func (noopMutationRoutes) Delete(http.ResponseWriter, *http.Request)  {}

type noopDashboardRoutes struct{}

func (noopDashboardRoutes) Get(http.ResponseWriter, *http.Request) {}

func TestSummaryEndpoint(t *testing.T) {
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
		noopWalletRoutes{},
		noopDebtRoutes{},
		noopMutationRoutes{},
		noopDashboardRoutes{},
		handler,
		middleware.NewAuthMiddleware(tokenManager),
	)

	req := httptest.NewRequest(http.MethodGet, "/summary?from=2026-04-01&to=2026-04-30", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 from /summary, got %d", res.Code)
	}
}
