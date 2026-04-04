package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
	"github.com/alpardfm/moneypath-api/internal/module/profile"
)

type integrationRepo struct {
	user *auth.User
}

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

type noopSummaryRoutes struct{}

func (noopSummaryRoutes) Get(http.ResponseWriter, *http.Request) {}

func (r *integrationRepo) CreateUser(ctx context.Context, user *auth.User) error {
	if r.user != nil && r.user.Email == user.Email {
		return auth.ErrEmailAlreadyUsed
	}
	if r.user != nil && r.user.Username == user.Username {
		return auth.ErrUsernameAlreadyUsed
	}

	user.ID = "user-1"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.user = user
	return nil
}

func (r *integrationRepo) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	return nil, auth.ErrUserNotFound
}

func (r *integrationRepo) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	return nil, auth.ErrUserNotFound
}

func (r *integrationRepo) GetUserByEmailOrUsername(ctx context.Context, value string) (*auth.User, error) {
	if r.user == nil {
		return nil, auth.ErrUserNotFound
	}
	if r.user.Email == value || r.user.Username == value {
		return r.user, nil
	}
	return nil, auth.ErrUserNotFound
}

func (r *integrationRepo) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	if r.user == nil || r.user.ID != userID {
		return nil, auth.ErrUserNotFound
	}
	return r.user, nil
}

func (r *integrationRepo) UpdateProfile(ctx context.Context, userID, email, username, fullName string) (*auth.User, error) {
	if r.user == nil || r.user.ID != userID {
		return nil, auth.ErrUserNotFound
	}

	r.user.Email = email
	r.user.Username = username
	r.user.FullName = fullName
	r.user.UpdatedAt = time.Now()
	return r.user, nil
}

func (r *integrationRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if r.user == nil || r.user.ID != userID {
		return auth.ErrUserNotFound
	}

	r.user.PasswordHash = passwordHash
	r.user.UpdatedAt = time.Now()
	return nil
}

func TestAuthAndProfileFlow(t *testing.T) {
	repo := &integrationRepo{}
	tokenManager := auth.NewTokenManager("secret")
	authHandler := auth.NewHandler(auth.NewService(repo, tokenManager))
	profileHandler := profile.NewHandler(profile.NewService(repo))
	router := NewRouter(
		slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		authHandler,
		profileHandler,
		noopWalletRoutes{},
		noopDebtRoutes{},
		noopMutationRoutes{},
		noopDashboardRoutes{},
		noopSummaryRoutes{},
		middleware.NewAuthMiddleware(tokenManager),
	)

	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"john@mail.com","username":"john","password":"password123","full_name":"John Doe"}`))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	router.ServeHTTP(registerRes, registerReq)
	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected 201 from register, got %d", registerRes.Code)
	}

	var registerPayload struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerRes.Body.Bytes(), &registerPayload); err != nil {
		t.Fatalf("unmarshal register payload: %v", err)
	}
	if registerPayload.Data.Token == "" {
		t.Fatal("expected register token")
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email_or_username":"john","password":"password123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	router.ServeHTTP(loginRes, loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d", loginRes.Code)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+registerPayload.Data.Token)
	meRes := httptest.NewRecorder()
	router.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from /me, got %d", meRes.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"email":"johnny@mail.com","username":"johnny","full_name":"Johnny Doe"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+registerPayload.Data.Token)
	updateRes := httptest.NewRecorder()
	router.ServeHTTP(updateRes, updateReq)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from PUT /me, got %d", updateRes.Code)
	}

	changePasswordReq := httptest.NewRequest(http.MethodPut, "/me/password", bytes.NewBufferString(`{"current_password":"password123","new_password":"new-password"}`))
	changePasswordReq.Header.Set("Content-Type", "application/json")
	changePasswordReq.Header.Set("Authorization", "Bearer "+registerPayload.Data.Token)
	changePasswordRes := httptest.NewRecorder()
	router.ServeHTTP(changePasswordRes, changePasswordReq)
	if changePasswordRes.Code != http.StatusOK {
		t.Fatalf("expected 200 from PUT /me/password, got %d", changePasswordRes.Code)
	}
}
