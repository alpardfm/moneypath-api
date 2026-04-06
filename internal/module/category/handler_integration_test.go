package category

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
	categories map[string]Category
}

func (r *integrationRepo) Create(ctx context.Context, item *Category) error {
	if r.categories == nil {
		r.categories = map[string]Category{}
	}
	item.ID = "category-1"
	item.IsActive = true
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt
	r.categories[item.ID] = *item
	return nil
}

func (r *integrationRepo) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	var items []Category
	for _, item := range r.categories {
		if item.UserID != userID || !item.IsActive || item.DeletedAt != nil {
			continue
		}
		if options.Type != "" && item.Type != options.Type {
			continue
		}
		items = append(items, item)
	}
	start := (options.Page - 1) * options.PageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + options.PageSize
	if end > len(items) {
		end = len(items)
	}
	return &ListResult{Items: items[start:end], TotalItems: len(items)}, nil
}

func (r *integrationRepo) GetByID(ctx context.Context, userID, categoryID string) (*Category, error) {
	item, ok := r.categories[categoryID]
	if !ok || item.UserID != userID {
		return nil, ErrCategoryNotFound
	}
	return &item, nil
}

func (r *integrationRepo) Update(ctx context.Context, userID, categoryID, name, categoryType string) (*Category, error) {
	item, ok := r.categories[categoryID]
	if !ok || item.UserID != userID {
		return nil, ErrCategoryNotFound
	}
	item.Name = name
	item.Type = categoryType
	item.UpdatedAt = time.Now()
	r.categories[categoryID] = item
	return &item, nil
}

func (r *integrationRepo) Inactivate(ctx context.Context, userID, categoryID string) error {
	item, ok := r.categories[categoryID]
	if !ok || item.UserID != userID {
		return ErrCategoryNotFound
	}
	now := time.Now()
	item.IsActive = false
	item.DeletedAt = &now
	item.UpdatedAt = now
	r.categories[categoryID] = item
	return nil
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

type noopNotificationRoutes struct{}

func (noopNotificationRoutes) Get(http.ResponseWriter, *http.Request) {}

func TestCategoryFlow(t *testing.T) {
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
		handler,
		noopMutationRoutes{},
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

	createReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":"Salary","type":"masuk"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRes := httptest.NewRecorder()
	router.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from create category, got %d", createRes.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/categories?type=masuk", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from list categories, got %d", listRes.Code)
	}

	var listPayload struct {
		Meta struct {
			TotalItems int `json:"total_items"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(listRes.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if listPayload.Meta.TotalItems != 1 {
		t.Fatalf("expected one category, got %+v", listPayload.Meta)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/categories/category-1", bytes.NewBufferString(`{"name":"Main Salary","type":"masuk"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from update category, got %d", updateRes.Code)
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/categories/category-1", nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailRes := httptest.NewRecorder()
	router.ServeHTTP(detailRes, detailReq)
	if detailRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from category detail, got %d", detailRes.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/category-1", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRes := httptest.NewRecorder()
	router.ServeHTTP(deleteRes, deleteReq)
	if deleteRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from category inactivate, got %d", deleteRes.Code)
	}
}
