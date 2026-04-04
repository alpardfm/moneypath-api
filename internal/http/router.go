package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/alpardfm/moneypath-api/internal/http/apidocs"
	appmiddleware "github.com/alpardfm/moneypath-api/internal/http/middleware"
)

// AuthRoutes exposes the auth handlers used by the router.
type AuthRoutes interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
}

// ProfileRoutes exposes the profile handlers used by the router.
type ProfileRoutes interface {
	GetMe(http.ResponseWriter, *http.Request)
	UpdateMe(http.ResponseWriter, *http.Request)
	ChangePassword(http.ResponseWriter, *http.Request)
}

// WalletRoutes exposes the wallet handlers used by the router.
type WalletRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	ListActive(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Inactivate(http.ResponseWriter, *http.Request)
}

// DebtRoutes exposes the debt handlers used by the router.
type DebtRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Inactivate(http.ResponseWriter, *http.Request)
}

// MutationRoutes exposes the mutation handlers used by the router.
type MutationRoutes interface {
	Create(http.ResponseWriter, *http.Request)
	List(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

// DashboardRoutes exposes the dashboard handlers used by the router.
type DashboardRoutes interface {
	Get(http.ResponseWriter, *http.Request)
}

// SummaryRoutes exposes the summary handlers used by the router.
type SummaryRoutes interface {
	Get(http.ResponseWriter, *http.Request)
}

// NewRouter creates the HTTP router used by the API.
func NewRouter(
	log *slog.Logger,
	healthHandler http.Handler,
	authRoutes AuthRoutes,
	profileRoutes ProfileRoutes,
	walletRoutes WalletRoutes,
	debtRoutes DebtRoutes,
	mutationRoutes MutationRoutes,
	dashboardRoutes DashboardRoutes,
	summaryRoutes SummaryRoutes,
	allowedOrigins []string,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(appmiddleware.RequestLogger(log))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Method(http.MethodGet, "/health", healthHandler)
	router.Get("/openapi.json", apidocs.OpenAPI)
	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusTemporaryRedirect)
	})
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/openapi.json"),
	))
	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", authRoutes.Register)
		r.Post("/login", authRoutes.Login)
	})
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/me", profileRoutes.GetMe)
		r.Put("/me", profileRoutes.UpdateMe)
		r.Put("/me/password", profileRoutes.ChangePassword)
		r.Route("/wallets", func(walletRouter chi.Router) {
			walletRouter.Post("/", walletRoutes.Create)
			walletRouter.Get("/", walletRoutes.ListActive)
			walletRouter.Get("/{walletID}", walletRoutes.GetByID)
			walletRouter.Put("/{walletID}", walletRoutes.Update)
			walletRouter.Delete("/{walletID}", walletRoutes.Inactivate)
		})
		r.Route("/debts", func(debtRouter chi.Router) {
			debtRouter.Post("/", debtRoutes.Create)
			debtRouter.Get("/", debtRoutes.List)
			debtRouter.Get("/{debtID}", debtRoutes.GetByID)
			debtRouter.Put("/{debtID}", debtRoutes.Update)
			debtRouter.Delete("/{debtID}", debtRoutes.Inactivate)
		})
		r.Route("/mutations", func(mutationRouter chi.Router) {
			mutationRouter.Post("/", mutationRoutes.Create)
			mutationRouter.Get("/", mutationRoutes.List)
			mutationRouter.Get("/{mutationID}", mutationRoutes.GetByID)
			mutationRouter.Put("/{mutationID}", mutationRoutes.Update)
			mutationRouter.Delete("/{mutationID}", mutationRoutes.Delete)
		})
		r.Get("/dashboard", dashboardRoutes.Get)
		r.Get("/summary", summaryRoutes.Get)
	})

	return router
}
