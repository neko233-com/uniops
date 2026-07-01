package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/server"
	"github.com/neko233/uniops/internal/store"
)

func main() {
	// Initialize database
	db, err := store.New("uniops.db")
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

	addr := ":8080"
	fmt.Printf("UniOps server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
