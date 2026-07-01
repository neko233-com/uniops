package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/server"
	"github.com/neko233/uniops/internal/store"
)

func main() {
	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "uniops.db"
	}
	db, err := store.New(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize admin user
	if err := db.InitAdmin(); err != nil {
		log.Fatal("Failed to initialize admin:", err)
	}

	// Initialize JWT manager
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	jwtManager := auth.NewJWTManager(secret)

	// Create router
	router := server.NewRouter(db, jwtManager)

	// Serve frontend static files
	frontendDir := filepath.Join(".", "web", "dist")
	if _, err := os.Stat(frontendDir); err == nil {
		fs := http.FileServer(http.Dir(frontendDir))
		router.Handle("/*", fs)
	}

	addr := ":6020"
	fmt.Printf("UniOps server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
