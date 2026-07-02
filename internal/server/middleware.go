package server

import (
	"net/http"
	"strings"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/oplog"
)

// OpLogMiddleware logs all API operations to daily files
func OpLogMiddleware(logger *oplog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip static files, health checks, and log reads
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "/api/health" || strings.HasPrefix(r.URL.Path, "/api/oplog") {
				next.ServeHTTP(w, r)
				return
			}

			wrapped := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(wrapped, r)

			user := "anonymous"
			userID := uint(0)
			if claims := auth.GetUserFromContext(r.Context()); claims != nil {
				user = claims.Username
				userID = claims.UserID
			}

			logger.Log(oplog.Entry{
				User:   user,
				UserID: userID,
				Action: r.Method + " " + r.URL.Path,
				Method: r.Method,
				Path:   r.URL.Path,
				Status: wrapped.status,
				IP:     r.RemoteAddr,
			})
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
