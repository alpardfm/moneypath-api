package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

// NewRouter creates the HTTP router used by the API.
func NewRouter(
	log *slog.Logger,
	healthHandler http.Handler,
	authRoutes AuthRoutes,
	profileRoutes ProfileRoutes,
	walletRoutes WalletRoutes,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(appmiddleware.RequestLogger(log))

	router.Method(http.MethodGet, "/health", healthHandler)
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
	})

	return router
}
