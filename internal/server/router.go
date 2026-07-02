package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/oplog"
	"github.com/neko233/uniops/internal/server/handlers"
	"github.com/neko233/uniops/internal/store"
)

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

func NewRouter(db *store.DB, jwtManager *auth.JWTManager, opLogger *oplog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)
	r.Use(OpLogMiddleware(opLogger))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		authHandler := handlers.NewAuthHandler(db, jwtManager)
		r.Post("/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware(jwtManager))

			r.Get("/auth/me", authHandler.Me)

			serverHandler := handlers.NewServerHandler(db)
			r.Route("/servers", func(r chi.Router) {
				r.Get("/", serverHandler.List)
				r.Post("/", serverHandler.Create)
				r.Get("/{id}", serverHandler.Get)
				r.Put("/{id}", serverHandler.Update)
				r.Delete("/{id}", serverHandler.Delete)
			})

			agentHandler := handlers.NewAgentHandler(db)
			r.Route("/agents", func(r chi.Router) {
				r.Get("/", agentHandler.List)
				r.Post("/", agentHandler.Create)
				r.Get("/{id}", agentHandler.Get)
				r.Put("/{id}", agentHandler.Update)
				r.Delete("/{id}", agentHandler.Delete)
				r.Post("/{id}/test", agentHandler.Test)
			})

			sshKeyHandler := handlers.NewSSHKeyHandler(db)
			r.Route("/sshkeys", func(r chi.Router) {
				r.Post("/generate", sshKeyHandler.Generate)
				r.Post("/deploy", sshKeyHandler.Deploy)
				r.Get("/server/{serverId}", sshKeyHandler.List)
				r.Get("/{id}", sshKeyHandler.Get)
			})

			fileManagerHandler := handlers.NewFileManagerHandler(db)
			r.Route("/files/{serverId}", func(r chi.Router) {
				r.Post("/list", fileManagerHandler.ListFiles)
				r.Get("/*", fileManagerHandler.Download)
				r.Put("/*", fileManagerHandler.Upload)
				r.Delete("/*", fileManagerHandler.Delete)
				r.Post("/mkdir", fileManagerHandler.Mkdir)
			})

			monitorHandler := handlers.NewMonitorHandler(db)
			r.Get("/monitor/{serverId}", monitorHandler.GetMetrics)

			agentChatHandler := handlers.NewAgentChatHandler(db)
			r.Post("/agent/chat", agentChatHandler.Chat)

			terminalHandler := handlers.NewTerminalHandler(db)
			r.Get("/ws/terminal/{serverId}", terminalHandler.Connect)

			auditHandler := handlers.NewAuditHandler(db)
			r.Route("/audit", func(r chi.Router) {
				r.Get("/sessions", auditHandler.ListSessions)
				r.Get("/sessions/{id}", auditHandler.GetSession)
				r.Get("/sessions/{id}/replay", auditHandler.GetReplay)
			})

			deployHandler := handlers.NewDeployHandler(db)
			r.Route("/deploy", func(r chi.Router) {
				r.Get("/", deployHandler.List)
				r.Post("/", deployHandler.Create)
				r.Get("/server/{serverId}", deployHandler.ListByServer)
				r.Get("/{id}", deployHandler.Get)
				r.Get("/{id}/ws", deployHandler.Watch)
			})
			r.Post("/deploy/exec", deployHandler.Exec)

			// Operation log routes
			opLogHandler := handlers.NewOpLogHandler(opLogger)
			r.Route("/oplog", func(r chi.Router) {
				r.Get("/dates", opLogHandler.ListDates)
				r.Get("/read", opLogHandler.ReadDate)
				r.Get("/search", opLogHandler.Search)
				r.Delete("/", opLogHandler.DeleteRange)
			})
		})
	})

	return r
}
