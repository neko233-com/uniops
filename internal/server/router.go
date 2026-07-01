package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/server/handlers"
	"github.com/neko233/uniops/internal/store"
)

// Prevent unused import error for gorilla/websocket (used in terminal.go)
var _ = websocket.Upgrader{}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewRouter(db *store.DB, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)

	// Public routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		// Auth routes
		authHandler := handlers.NewAuthHandler(db, jwtManager)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(jwtManager))

			r.Get("/auth/me", authHandler.Me)

			// Server routes
			serverHandler := handlers.NewServerHandler(db)
			r.Route("/servers", func(r chi.Router) {
				r.Get("/", serverHandler.List)
				r.Post("/", serverHandler.Create)
				r.Get("/{id}", serverHandler.Get)
				r.Put("/{id}", serverHandler.Update)
				r.Delete("/{id}", serverHandler.Delete)
			})

			// Agent routes
			agentHandler := handlers.NewAgentHandler(db)
			r.Route("/agents", func(r chi.Router) {
				r.Get("/", agentHandler.List)
				r.Post("/", agentHandler.Create)
				r.Delete("/{id}", agentHandler.Delete)
			})

			// Terminal WebSocket
			terminalHandler := handlers.NewTerminalHandler(db)
			r.Get("/ws/terminal/{serverId}", terminalHandler.Connect)
		})
	})

	return r
}
