package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/alpardfm/moneypath-api/internal/http/response"
	"github.com/alpardfm/moneypath-api/internal/module/auth"
)

type authContextKey string

const userIDContextKey authContextKey = "auth_user_id"

// NewAuthMiddleware validates bearer tokens and stores the user id in request context.
func NewAuthMiddleware(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}

			claims, err := tokens.Parse(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthUserID retrieves the authenticated user id from context.
func AuthUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}
