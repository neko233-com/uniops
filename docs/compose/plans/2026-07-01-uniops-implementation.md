# UniOps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an Agent-First bastion host platform with Go backend, SQLite, and React desktop-style UI

**Architecture:** Go HTTP server with WebSocket support, SQLite for persistence, React frontend with desktop-like UI (windows, taskbar, sidebar), xterm.js for terminal, Agent integration via API

**Tech Stack:** Go 1.26, Chi Router, GORM, SQLite, React 19, Vite, TypeScript, TailwindCSS, Radix UI, xterm.js

## Global Constraints

- Go 1.26 minimum
- SQLite for all data storage
- React 19 + Vite + TypeScript for frontend
- All API keys encrypted at rest
- JWT authentication with refresh tokens
- RBAC: admin, operator, viewer roles

---

## Task 1: Project Scaffolding

**Covers:** [S2, S3]

**Files:**
- Create: `go.mod`
- Create: `cmd/uniops/main.go`
- Create: `internal/server/router.go`
- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`

**Interfaces:**
- Produces: Basic project structure that compiles and runs

- [ ] **Step 1: Initialize Go module**

```bash
cd D:/Code/neko233-Projects/uniops
go mod init github.com/neko233/uniops
```

- [ ] **Step 2: Install Go dependencies**

```bash
go get github.com/go-chi/chi/v5
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

- [ ] **Step 3: Create main.go**

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/neko233/uniops/internal/server"
)

func main() {
	router := server.NewRouter()
	
	addr := ":8080"
	fmt.Printf("UniOps server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
```

- [ ] **Step 4: Create router.go**

```go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	
	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
	})
	
	return r
}
```

- [ ] **Step 5: Initialize React project**

```bash
cd D:/Code/neko233-Projects/uniops
npm create vite@latest web -- --template react-ts
cd web
npm install
```

- [ ] **Step 6: Install frontend dependencies**

```bash
cd D:/Code/neko233-Projects/uniops/web
npm install @radix-ui/react-icons @radix-ui/react-dialog @radix-ui/react-dropdown-menu
npm install tailwindcss @tailwindcss/vite
npm install xterm xterm-addon-fit xterm-addon-web-links
```

- [ ] **Step 7: Create basic App.tsx**

```tsx
import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <h1 className="text-2xl p-4">UniOps</h1>
      <button 
        className="bg-blue-500 px-4 py-2 rounded"
        onClick={() => setCount(c => c + 1)}
      >
        Count: {count}
      </button>
    </div>
  )
}

export default App
```

- [ ] **Step 8: Verify Go builds**

```bash
cd D:/Code/neko233-Projects/uniops
go build ./cmd/uniops
```

Expected: Binary created without errors

- [ ] **Step 9: Verify React builds**

```bash
cd D:/Code/neko233-Projects/uniops/web
npm run build
```

Expected: Build successful

- [ ] **Step 10: Commit**

```bash
git init
git add .
git commit -m "feat: project scaffolding with Go backend and React frontend"
```

---

## Task 2: SQLite Store Layer

**Covers:** [S3]

**Files:**
- Create: `internal/store/sqlite.go`
- Create: `internal/model/user.go`
- Create: `internal/model/server.go`
- Create: `internal/model/agent.go`
- Create: `internal/model/session.go`
- Create: `internal/model/command.go`

**Interfaces:**
- Consumes: None
- Produces: `store.DB` with auto-migration, CRUD methods for all models

- [ ] **Step 1: Create model/user.go**

```go
package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex" json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // admin, operator, viewer
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Create model/server.go**

```go
package model

import "time"

type Server struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `gorm:"default:22" json:"port"`
	Username  string    `json:"username"`
	AuthType  string    `json:"auth_type"` // password, key
	AuthData  string    `json:"auth_data"` // encrypted
	AgentID   uint      `json:"agent_id"`
	GroupID   uint      `json:"group_id"`
	Status    string    `gorm:"default:offline" json:"status"` // online, offline
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 3: Create model/agent.go**

```go
package model

import "time"

type Agent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // claude, openai, custom
	Endpoint  string    `json:"endpoint"`
	APIKey    string    `json:"api_key"` // encrypted
	Config    string    `json:"config"`  // JSON
	Status    string    `gorm:"default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 4: Create model/session.go**

```go
package model

import "time"

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	ServerID  uint      `json:"server_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Status    string    `gorm:"default:active" json:"status"` // active, closed
	Replay    string    `json:"replay"` // base64 encoded
}
```

- [ ] **Step 5: Create model/command.go**

```go
package model

import "time"

type Command struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID uint      `json:"session_id"`
	Command   string    `json:"command"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}
```

- [ ] **Step 6: Create store/sqlite.go**

```go
package store

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
	"github.com/neko233/uniops/internal/model"
)

type DB struct {
	*gorm.DB
}

func New(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	
	// Auto migrate
	err = db.AutoMigrate(
		&model.User{},
		&model.Server{},
		&model.Agent{},
		&model.Session{},
		&model.Command{},
	)
	if err != nil {
		return nil, err
	}
	
	return &DB{db}, nil
}

func (db *DB) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (db *DB) CreateUser(user *model.User) error {
	return db.Create(user).Error
}

func (db *DB) GetServers() ([]model.Server, error) {
	var servers []model.Server
	err := db.Find(&servers).Error
	return servers, err
}

func (db *DB) GetServer(id uint) (*model.Server, error) {
	var server model.Server
	err := db.First(&server, id).Error
	return &server, err
}

func (db *DB) CreateServer(server *model.Server) error {
	return db.Create(server).Error
}

func (db *DB) UpdateServer(server *model.Server) error {
	return db.Save(server).Error
}

func (db *DB) DeleteServer(id uint) error {
	return db.Delete(&model.Server{}, id).Error
}

func (db *DB) GetAgents() ([]model.Agent, error) {
	var agents []model.Agent
	err := db.Find(&agents).Error
	return agents, err
}

func (db *DB) CreateAgent(agent *model.Agent) error {
	return db.Create(agent).Error
}

func (db *DB) GetSessions() ([]model.Session, error) {
	var sessions []model.Session
	err := db.Find(&sessions).Error
	return sessions, err
}

func (db *DB) CreateSession(session *model.Session) error {
	return db.Create(session).Error
}

func (db *DB) GetSessionByID(id uint) (*model.Session, error) {
	var session model.Session
	err := db.First(&session, id).Error
	return &session, err
}

func (db *DB) UpdateSession(session *model.Session) error {
	return db.Save(session).Error
}

func (db *DB) CreateCommand(cmd *model.Command) error {
	return db.Create(cmd).Error
}

func (db *DB) GetCommandsBySession(sessionID uint) ([]model.Command, error) {
	var commands []model.Command
	err := db.Where("session_id = ?", sessionID).Find(&commands).Error
	return commands, err
}
```

- [ ] **Step 7: Create init admin user function**

```go
package store

import (
	"golang.org/x/crypto/bcrypt"
	
	"github.com/neko233/uniops/internal/model"
)

func (db *DB) InitAdmin() error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	admin := &model.User{
		Username: "admin",
		Password: string(hash),
		Role:     "admin",
	}
	return db.Create(admin).Error
}
```

- [ ] **Step 8: Commit**

```bash
git add internal/store/ internal/model/
git commit -m "feat: add SQLite store layer with data models"
```

---

## Task 3: JWT Authentication

**Covers:** [S3, S6]

**Files:**
- Create: `internal/auth/jwt.go`
- Create: `internal/auth/middleware.go`
- Create: `internal/server/handlers/auth.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB` from Task 2
- Produces: `auth.JWTManager`, auth middleware, login/logout endpoints

- [ ] **Step 1: Create auth/jwt.go**

```go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (m *JWTManager) GenerateAccessToken(userID uint, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) GenerateRefreshToken(userID uint, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	
	return claims, nil
}
```

- [ ] **Step 2: Create auth/middleware.go**

```go
package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			
			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(UserContextKey).(*Claims)
	return claims
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
```

- [ ] **Step 3: Create handlers/auth.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/store"
)

type AuthHandler struct {
	db         *store.DB
	jwtManager *auth.JWTManager
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func NewAuthHandler(db *store.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{db: db, jwtManager: jwtManager}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	
	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.Role = user.Role
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
```

- [ ] **Step 4: Update router.go**

```go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/server/handlers"
	"github.com/neko233/uniops/internal/store"
)

func NewRouter(db *store.DB, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware CORS)
	
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
		})
	})
	
	return r
}
```

- [ ] **Step 5: Update main.go**

```go
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
```

- [ ] **Step 6: Test login endpoint**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

Expected: Returns access_token and refresh_token

- [ ] **Step 7: Commit**

```bash
git add internal/auth/ internal/server/handlers/
git commit -m "feat: add JWT authentication with login endpoint"
```

---

## Task 4: Server Management API

**Covers:** [S3, S5]

**Files:**
- Create: `internal/server/handlers/server.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`, `auth.JWTManager`
- Produces: CRUD endpoints for server management

- [ ] **Step 1: Create handlers/server.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/store"
)

type ServerHandler struct {
	db *store.DB
}

type CreateServerRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	AuthType string `json:"auth_type"`
	AuthData string `json:"auth_data"`
	AgentID  uint   `json:"agent_id"`
	GroupID  uint   `json:"group_id"`
}

func NewServerHandler(db *store.DB) *ServerHandler {
	return &ServerHandler{db: db}
}

func (h *ServerHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.db.GetServers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func (h *ServerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	
	server, err := h.db.GetServer(uint(id))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	
	server := &model.Server{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		AuthType: req.AuthType,
		AuthData: req.AuthData,
		AgentID:  req.AgentID,
		GroupID:  req.GroupID,
	}
	
	if err := h.db.CreateServer(server); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	
	server, err := h.db.GetServer(uint(id))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	
	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	
	server.Name = req.Name
	server.Host = req.Host
	server.Port = req.Port
	server.Username = req.Username
	server.AuthType = req.AuthType
	server.AuthData = req.AuthData
	server.AgentID = req.AgentID
	server.GroupID = req.GroupID
	
	if err := h.db.UpdateServer(server); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	
	if err := h.db.DeleteServer(uint(id)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: Update router.go with server routes**

```go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/server/handlers"
	"github.com/neko233/uniops/internal/store"
)

func NewRouter(db *store.DB, jwtManager *auth.JWTManager) *chi.Mux {
	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.CORS)
	
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
		})
	})
	
	return r
}
```

- [ ] **Step 3: Test server CRUD**

```bash
# Create server
curl -X POST http://localhost:8080/api/servers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Web-1","host":"192.168.1.100","port":22,"username":"root"}'

# List servers
curl http://localhost:8080/api/servers \
  -H "Authorization: Bearer <token>"
```

- [ ] **Step 4: Commit**

```bash
git add internal/server/handlers/server.go internal/server/router.go
git commit -m "feat: add server management API endpoints"
```

---

## Task 5: Agent Management API

**Covers:** [S3, S5]

**Files:**
- Create: `internal/server/handlers/agent.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`
- Produces: CRUD endpoints for agent management

- [ ] **Step 1: Create handlers/agent.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/store"
)

type AgentHandler struct {
	db *store.DB
}

type CreateAgentRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
	Config   string `json:"config"`
}

func NewAgentHandler(db *store.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.db.GetAgents()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	
	agent := &model.Agent{
		Name:     req.Name,
		Type:     req.Type,
		Endpoint: req.Endpoint,
		APIKey:   req.APIKey,
		Config:   req.Config,
	}
	
	if err := h.db.CreateAgent(agent); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	
	if err := h.db.DeleteAgent(uint(id)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: Update router.go with agent routes**

Add to the protected routes group:

```go
// Agent routes
agentHandler := handlers.NewAgentHandler(db)
r.Route("/agents", func(r chi.Router) {
	r.Get("/", agentHandler.List)
	r.Post("/", agentHandler.Create)
	r.Delete("/{id}", agentHandler.Delete)
})
```

- [ ] **Step 3: Commit**

```bash
git add internal/server/handlers/agent.go
git commit -m "feat: add agent management API endpoints"
```

---

## Task 6: WebSocket Terminal

**Covers:** [S3, S5]

**Files:**
- Create: `internal/ssh/proxy.go`
- Create: `internal/server/handlers/terminal.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`
- Produces: WebSocket endpoint for terminal access

- [ ] **Step 1: Create ssh/proxy.go**

```go
package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	mu      sync.Mutex
}

func NewSession(host string, port int, username, authType, authData string) (*Session, error) {
	var authMethod ssh.AuthMethod
	
	switch authType {
	case "password":
		authMethod = ssh.Password(authData)
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(authData))
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		authMethod = ssh.PublicKeys(signer)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}
	
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}
	
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}
	
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	
	if err := session.RequestPty("xterm-256color", 40, 100, modes); err != nil {
		session.Close()
		client.Close()
		return nil, err
	}
	
	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, err
	}
	
	return &Session{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}, nil
}

func (s *Session) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.stdin.Write(data)
	return err
}

func (s *Session) Read(buf []byte) (int, error) {
	return s.stdout.Read(buf)
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session.Close()
	return s.client.Close()
}

func (s *Session) Resize(width, height int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.session.WindowChange(height, width)
}
```

- [ ] **Step 2: Create handlers/terminal.go**

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/ssh"
	"github.com/neko233/uniops/internal/store"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalHandler struct {
	db *store.DB
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func NewTerminalHandler(db *store.DB) *TerminalHandler {
	return &TerminalHandler{db: db}
}

func (h *TerminalHandler) Connect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	if err != nil {
		http.Error(w, "invalid server id", http.StatusBadRequest)
		return
	}
	
	server, err := h.db.GetServer(uint(id))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	
	session, err := ssh.NewSession(
		server.Host,
		server.Port,
		server.Username,
		server.AuthType,
		server.AuthData,
	)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(
			`{"type":"error","data":"`+err.Error()+`"}`,
		))
		return
	}
	defer session.Close()
	
	var mu sync.Mutex
	
	// Read from WebSocket and write to SSH
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			
			var msg TerminalMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}
			
			switch msg.Type {
			case "input":
				mu.Lock()
				session.Write([]byte(msg.Data))
				mu.Unlock()
			case "resize":
				var dims struct {
					Width  int `json:"width"`
					Height int `json:"height"`
				}
				if err := json.Unmarshal([]byte(msg.Data), &dims); err == nil {
					session.Resize(dims.Width, dims.Height)
				}
			}
		}
	}()
	
	// Read from SSH and write to WebSocket
	buf := make([]byte, 1024)
	for {
		n, err := session.Read(buf)
		if err != nil {
			break
		}
		
		msg := TerminalMessage{
			Type: "output",
			Data: string(buf[:n]),
		}
		
		mu.Lock()
		conn.WriteJSON(msg)
		mu.Unlock()
	}
}
```

- [ ] **Step 3: Update router.go with WebSocket endpoint**

Add to protected routes:

```go
import "github.com/gorilla/websocket"

// Terminal WebSocket
terminalHandler := handlers.NewTerminalHandler(db)
r.Get("/ws/terminal/{serverId}", terminalHandler.Connect)
```

- [ ] **Step 4: Add gorilla/websocket dependency**

```bash
go get github.com/gorilla/websocket
```

- [ ] **Step 5: Commit**

```bash
go get github.com/gorilla/websocket
git add internal/ssh/ internal/server/handlers/terminal.go
git commit -m "feat: add WebSocket terminal with SSH proxy"
```

---

## Task 7: Session Recording

**Covers:** [S3, S5]

**Files:**
- Create: `internal/audit/recorder.go`
- Modify: `internal/server/handlers/terminal.go`

**Interfaces:**
- Consumes: `store.DB`, SSH session data
- Produces: Session recording for audit

- [ ] **Step 1: Create audit/recorder.go**

```go
package audit

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Recorder struct {
	sessionID uint
	commands  []RecordEntry
}

type RecordEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // input, output
	Data      string    `json:"data"`
}

func NewRecorder(sessionID uint) *Recorder {
	return &Recorder{sessionID: sessionID}
}

func (r *Recorder) RecordInput(data string) {
	r.commands = append(r.commands, RecordEntry{
		Timestamp: time.Now(),
		Type:      "input",
		Data:      data,
	})
}

func (r *Recorder) RecordOutput(data string) {
	r.commands = append(r.commands, RecordEntry{
		Timestamp: time.Now(),
		Type:      "output",
		Data:      data,
	})
}

func (r *Recorder) GetReplay() string {
	data, _ := json.Marshal(r.commands)
	return base64.StdEncoding.EncodeToString(data)
}
```

- [ ] **Step 2: Update terminal handler to record sessions**

Modify `handlers/terminal.go` to use recorder.

- [ ] **Step 3: Commit**

```bash
git add internal/audit/
git commit -m "feat: add session recording for audit"
```

---

## Task 8: React Desktop Layout

**Covers:** [S4]

**Files:**
- Create: `web/src/components/Desktop.tsx`
- Create: `web/src/components/Taskbar.tsx`
- Create: `web/src/components/Sidebar.tsx`
- Create: `web/src/components/Window.tsx`
- Modify: `web/src/App.tsx`

**Interfaces:**
- Consumes: None
- Produces: Desktop-style UI layout

- [ ] **Step 1: Create components/Desktop.tsx**

```tsx
import { useState } from 'react'
import { Taskbar } from './Taskbar'
import { Sidebar } from './Sidebar'
import { Window } from './Window'

interface App {
  id: string
  title: string
  icon: string
  component: React.ComponentType
}

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '🖥️', component: () => <div>Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <div>Files</div> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <div>Monitor</div> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <div>Agent</div> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div>Audit</div> },
]

export function Desktop() {
  const [activeApp, setActiveApp] = useState<string | null>(null)
  
  const ActiveComponent = apps.find(a => a.id === activeApp)?.component
  
  return (
    <div className="h-screen flex flex-col bg-gray-900">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} />
        <main className="flex-1 overflow-auto">
          {ActiveComponent && <ActiveComponent />}
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
```

- [ ] **Step 2: Create components/Taskbar.tsx**

```tsx
interface TaskbarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  activeApp: string | null
  onSelect: (id: string) => void
}

export function Taskbar({ apps, activeApp, onSelect }: TaskbarProps) {
  return (
    <div className="h-12 bg-gray-800 border-t border-gray-700 flex items-center px-4">
      <div className="flex gap-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className={`px-3 py-1 rounded text-sm ${
              activeApp === app.id
                ? 'bg-blue-600 text-white'
                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
            }`}
          >
            {app.icon} {app.title}
          </button>
        ))}
      </div>
      <div className="ml-auto text-gray-400 text-sm">
        UniOps
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Create components/Sidebar.tsx**

```tsx
interface SidebarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  onSelect: (id: string) => void
}

export function Sidebar({ apps, onSelect }: SidebarProps) {
  return (
    <div className="w-48 bg-gray-800 border-r border-gray-700 p-4">
      <h2 className="text-white font-bold mb-4">UniOps</h2>
      <nav className="space-y-1">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className="w-full text-left px-3 py-2 rounded text-gray-300 hover:bg-gray-700"
          >
            {app.icon} {app.title}
          </button>
        ))}
      </nav>
    </div>
  )
}
```

- [ ] **Step 4: Create components/Window.tsx**

```tsx
interface WindowProps {
  title: string
  children: React.ReactNode
}

export function Window({ title, children }: WindowProps) {
  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden h-full">
      <div className="bg-gray-700 px-4 py-2 flex items-center justify-between">
        <span className="text-white font-medium">{title}</span>
        <div className="flex gap-2">
          <button className="w-3 h-3 rounded-full bg-yellow-500" />
          <button className="w-3 h-3 rounded-full bg-green-500" />
          <button className="w-3 h-3 rounded-full bg-red-500" />
        </div>
      </div>
      <div className="p-4">
        {children}
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Update App.tsx**

```tsx
import { Desktop } from './components/Desktop'

function App() {
  return <Desktop />
}

export default App
```

- [ ] **Step 6: Commit**

```bash
git add web/src/components/
git commit -m "feat: add desktop-style UI layout"
```

---

## Task 9: Terminal Component

**Covers:** [S4, S5]

**Files:**
- Create: `web/src/components/Terminal.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: WebSocket from Task 6
- Produces: Interactive terminal component

- [ ] **Step 1: Create components/Terminal.tsx**

```tsx
import { useEffect, useRef } from 'react'
import { Terminal as XTerminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

interface TerminalProps {
  serverId: number
}

export function Terminal({ serverId }: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<XTerminal | null>(null)
  
  useEffect(() => {
    if (!containerRef.current) return
    
    const term = new XTerminal({
      theme: {
        background: '#1a1a2e',
        foreground: '#eaeaea',
      },
    })
    
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(containerRef.current)
    fitAddon.fit()
    
    termRef.current = term
    
    const ws = new WebSocket(`ws://localhost:8080/ws/terminal/${serverId}`)
    
    ws.onopen = () => {
      ws.send(JSON.stringify({
        type: 'resize',
        data: JSON.stringify({ width: term.cols, height: term.rows }),
      }))
    }
    
    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      if (msg.type === 'output') {
        term.write(msg.data)
      }
    }
    
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', data }))
      }
    })
    
    const resizeHandler = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          data: JSON.stringify({ width: term.cols, height: term.rows }),
        }))
      }
    }
    
    window.addEventListener('resize', resizeHandler)
    
    return () => {
      window.removeEventListener('resize', resizeHandler)
      ws.close()
      term.dispose()
    }
  }, [serverId])
  
  return <div ref={containerRef} className="h-full w-full" />
}
```

- [ ] **Step 2: Update Desktop.tsx to use Terminal**

- [ ] **Step 3: Commit**

```bash
git add web/src/components/Terminal.tsx
git commit -m "feat: add xterm.js terminal component"
```

---

## Task 10: Build & Test

**Covers:** [S2, S7]

**Files:**
- Modify: Various

**Interfaces:**
- Consumes: All previous tasks
- Produces: Working application

- [ ] **Step 1: Build Go backend**

```bash
cd D:/Code/neko233-Projects/uniops
go build -o uniops.exe ./cmd/uniops
```

- [ ] **Step 2: Build React frontend**

```bash
cd D:/Code/neko233-Projects/uniops/web
npm run build
```

- [ ] **Step 3: Start server and test**

```bash
./uniops.exe
# Open http://localhost:8080
```

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "feat: initial working version of UniOps"
```
