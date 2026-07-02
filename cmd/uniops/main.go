package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/oplog"
	"github.com/neko233/uniops/internal/server"
	"github.com/neko233/uniops/internal/store"
	"github.com/neko233/uniops/web"
)

var version = "dev"

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "uniops.db"
	}
	db, err := store.New(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	if err := db.InitAdmin(); err != nil {
		log.Fatal("Failed to initialize admin:", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	jwtManager := auth.NewJWTManager(secret)

	opLogger := oplog.New("logs")
	router := server.NewRouter(db, jwtManager, opLogger)

	// Serve frontend: embedded FS with disk fallback for dev
	var frontendFS http.FileSystem
	if _, err := os.Stat("web/dist/index.html"); err == nil {
		frontendFS = http.Dir("web/dist")
		log.Println("Serving frontend from disk (web/dist)")
	} else {
		sub, err := fs.Sub(web.DistFS, "dist")
		if err != nil {
			log.Fatal("Failed to create sub FS:", err)
		}
		frontendFS = http.FS(sub)
		log.Println("Serving frontend from embedded FS")
	}

	// SPA fallback: serve index.html for non-API, non-static routes
	router.Handle("/*", spaHandler(frontendFS))

	addr := ":6020"
	fmt.Printf("UniOps %s starting on %s\n", version, addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// spaHandler serves static files with SPA fallback to index.html
func spaHandler(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API and WebSocket paths
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			http.NotFound(w, r)
			return
		}
		// Try serving the file directly
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := root.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
