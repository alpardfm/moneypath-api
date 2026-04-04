package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// RequestLogger logs the basic HTTP request lifecycle.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			userID, _ := AuthUserID(r.Context())

			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode,
				"request_id", chimiddleware.GetReqID(r.Context()),
				"user_id", userID,
				"duration", time.Since(start).String(),
			)
		})
	}
}
