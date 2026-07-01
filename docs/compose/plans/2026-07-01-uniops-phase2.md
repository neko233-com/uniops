# UniOps Phase 2: Enhanced Features Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add SSH fingerprint auto-deployment, ROG dark UI, pure Go SQLite, GitHub CI/CD, Docker testing, and GitHub Pages docs

**Architecture:** Extend existing Go backend with SSH key management, replace CGO SQLite with pure Go driver, add CI/CD pipelines, Docker test environment, and ROG-themed React UI

**Tech Stack:** Go 1.26, modernc.org/sqlite, React 19, Vite, TailwindCSS, Docker, GitHub Actions

## Global Constraints

- Go 1.26 minimum
- Pure Go SQLite (modernc.org/sqlite) - NO CGO
- ROG (Republic of Gamers) dark aesthetic for UI
- GitHub Actions for CI/CD binary packaging
- Docker for local testing only (NOT in CI)
- GitHub Pages for documentation

---

## Task 1: Replace CGO SQLite with Pure Go Driver

**Covers:** [S2, S6]

**Files:**
- Modify: `go.mod`
- Modify: `internal/store/sqlite.go`
- Modify: `cmd/uniops/main.go`

**Interfaces:**
- Consumes: Existing store layer
- Produces: CGO-free SQLite driver

- [ ] **Step 1: Update go.mod to use modernc.org/sqlite**

```bash
cd D:/Code/neko233-Projects/uniops
go get modernc.org/sqlite
go get github.com/glebarez/sqlite
```

- [ ] **Step 2: Update store/sqlite.go driver import**

```go
package store

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/glebarez/sqlite"
	
	"github.com/neko233/uniops/internal/model"
)

type DB struct {
	*gorm.DB
}

func New(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
		&model.SSHKey{},
	)
	if err != nil {
		return nil, err
	}
	
	return &DB{db}, nil
}

// ... rest of CRUD methods remain the same
```

- [ ] **Step 3: Remove old go-sqlite3 dependency**

```bash
go mod tidy
```

- [ ] **Step 4: Verify build without CGO**

```bash
CGO_ENABLED=0 go build -o uniops.exe ./cmd/uniops
```

Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/store/sqlite.go
git commit -m "feat: replace CGO sqlite with pure Go driver (modernc.org/sqlite)"
```

---

## Task 2: SSH Key Management Model & Store

**Covers:** [S3, S5]

**Files:**
- Create: `internal/model/sshkey.go`
- Modify: `internal/store/sqlite.go`

**Interfaces:**
- Consumes: Existing store
- Produces: SSHKey model and CRUD methods

- [ ] **Step 1: Create model/sshkey.go**

```go
package model

import "time"

type SSHKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `gorm:"type:text" json:"public_key"`
	PrivateKey string   `gorm:"type:text" json:"private_key"`
	Fingerprint string  `json:"fingerprint"`
	ServerID  uint      `json:"server_id"`
	Status    string    `gorm:"default:active" json:"status"` // active, deployed, pending
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Add SSHKey CRUD methods to store/sqlite.go**

```go
func (db *DB) GetSSHKeysByServer(serverID uint) ([]model.SSHKey, error) {
	var keys []model.SSHKey
	err := db.Where("server_id = ?", serverID).Find(&keys).Error
	return keys, err
}

func (db *DB) CreateSSHKey(key *model.SSHKey) error {
	return db.Create(key).Error
}

func (db *DB) GetSSHKeyByFingerprint(fingerprint string) (*model.SSHKey, error) {
	var key model.SSHKey
	err := db.Where("fingerprint = ?", fingerprint).First(&key).Error
	return &key, err
}

func (db *DB) UpdateSSHKey(key *model.SSHKey) error {
	return db.Save(key).Error
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/model/sshkey.go internal/store/sqlite.go
git commit -m "feat: add SSH key model and store methods"
```

---

## Task 3: SSH Key Auto-Deployment Service

**Covers:** [S3, S5, S6]

**Files:**
- Create: `internal/ssh/keymanager.go`
- Create: `internal/ssh/deployer.go`

**Interfaces:**
- Consumes: `store.DB`, `model.SSHKey`, `model.Server`
- Produces: Auto-deploy SSH keys to target servers

- [ ] **Step 1: Create ssh/keymanager.go**

```go
package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/ssh"
)

type KeyManager struct{}

func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

func (km *KeyManager) GenerateKeyPair(bits int) (privateKey string, publicKey string, fingerprint string, err error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Marshal private key
	privateKeyBytes, err := ssh.MarshalPrivateKey(key, "")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyPEM, err := ssh.MarshalOpenSSHPrivateKey(key, "")
	if err != nil {
		// Fallback to standard format
		import "encoding/pem"
		privateKeyPEM = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyDER,
		})
	}

	// Generate public key
	pubKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pubKey)

	// Calculate fingerprint
	hash := sha256.Sum256(pubKey.Marshal())
	fingerprint = hex.EncodeToString(hash[:])

	return string(privateKeyPEM), string(publicKeyBytes), fingerprint, nil
}

func (km *KeyManager) GetFingerprint(privateKey string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKey := signer.PublicKey()
	hash := sha256.Sum256(pubKey.Marshal())
	return hex.EncodeToString(hash[:]), nil
}
```

- [ ] **Step 2: Create ssh/deployer.go**

```go
package ssh

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Deployer struct{}

func NewDeployer() *Deployer {
	return &Deployer{}
}

func (d *Deployer) DeployKey(host string, port int, username, password, publicKey string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create .ssh directory and authorized_keys file
	commands := []string{
		"mkdir -p ~/.ssh",
		"chmod 700 ~/.ssh",
		fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", strings.TrimSpace(publicKey)),
		"chmod 600 ~/.ssh/authorized_keys",
	}

	for _, cmd := range commands {
		if err := session.Run(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %w", cmd, err)
		}
	}

	return nil
}

func (d *Deployer) TestKeyAuth(host string, port int, username, privateKey string) error {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("key auth failed: %w", err)
	}
	defer client.Close()

	return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/ssh/keymanager.go internal/ssh/deployer.go
git commit -m "feat: add SSH key manager and auto-deployer"
```

---

## Task 4: Agent Configuration API

**Covers:** [S3, S5]

**Files:**
- Modify: `internal/server/handlers/agent.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`
- Produces: Agent config endpoints for Claude/Codex/Custom

- [ ] **Step 1: Update agent handler with config validation**

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
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // claude, openai, custom
	Endpoint string                 `json:"endpoint"`
	APIKey   string                 `json:"api_key"`
	Config   map[string]interface{} `json:"config"`
}

type AgentConfigResponse struct {
	ID       uint                   `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Endpoint string                 `json:"endpoint"`
	Config   map[string]interface{} `json:"config"`
	Status   string                 `json:"status"`
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

	var responses []AgentConfigResponse
	for _, agent := range agents {
		var config map[string]interface{}
		json.Unmarshal([]byte(agent.Config), &config)
		responses = append(responses, AgentConfigResponse{
			ID:       agent.ID,
			Name:     agent.Name,
			Type:     agent.Type,
			Endpoint: agent.Endpoint,
			Config:   config,
			Status:   agent.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate agent type
	validTypes := map[string]bool{"claude": true, "openai": true, "custom": true}
	if !validTypes[req.Type] {
		http.Error(w, "invalid agent type", http.StatusBadRequest)
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	agent := &model.Agent{
		Name:     req.Name,
		Type:     req.Type,
		Endpoint: req.Endpoint,
		APIKey:   req.APIKey,
		Config:   string(configJSON),
	}

	if err := h.db.CreateAgent(agent); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	agent, err := h.db.GetAgent(uint(id))
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	var config map[string]interface{}
	json.Unmarshal([]byte(agent.Config), &config)

	resp := AgentConfigResponse{
		ID:       agent.ID,
		Name:     agent.Name,
		Type:     agent.Type,
		Endpoint: agent.Endpoint,
		Config:   config,
		Status:   agent.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	agent, err := h.db.GetAgent(uint(id))
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	var req CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agent.Name = req.Name
	agent.Type = req.Type
	agent.Endpoint = req.Endpoint
	agent.APIKey = req.APIKey
	configJSON, _ := json.Marshal(req.Config)
	agent.Config = string(configJSON)

	if err := h.db.UpdateAgent(agent); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

func (h *AgentHandler) Test(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	agent, err := h.db.GetAgent(uint(id))
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	// TODO: Implement actual agent test based on type
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"agent":  agent.Name,
	})
}
```

- [ ] **Step 2: Update router with full agent routes**

```go
// Agent routes
agentHandler := handlers.NewAgentHandler(db)
r.Route("/agents", func(r chi.Router) {
	r.Get("/", agentHandler.List)
	r.Post("/", agentHandler.Create)
	r.Get("/{id}", agentHandler.Get)
	r.Put("/{id}", agentHandler.Update)
	r.Delete("/{id}", agentHandler.Delete)
	r.Post("/{id}/test", agentHandler.Test)
})
```

- [ ] **Step 3: Commit**

```bash
git add internal/server/handlers/agent.go internal/server/router.go
git commit -m "feat: enhance agent management with config and test endpoints"
```

---

## Task 5: ROG Dark UI Theme

**Covers:** [S4]

**Files:**
- Create: `web/src/styles/rog-theme.css`
- Modify: `web/src/index.css`
- Modify: `web/src/components/Desktop.tsx`
- Modify: `web/src/components/Taskbar.tsx`
- Modify: `web/src/components/Sidebar.tsx`

**Interfaces:**
- Consumes: None
- Produces: ROG-themed dark UI

- [ ] **Step 1: Create rog-theme.css**

```css
/* ROG (Republic of Gamers) Dark Theme */
:root {
  --rog-red: #ff0033;
  --rog-red-dark: #cc0029;
  --rog-black: #0a0a0a;
  --rog-dark: #121212;
  --rog-gray: #1a1a1a;
  --rog-gray-light: #2a2a2a;
  --rog-gray-lighter: #3a3a3a;
  --rog-text: #e0e0e0;
  --rog-text-dim: #888888;
  --rog-accent: #ff0033;
  --rog-accent-glow: rgba(255, 0, 51, 0.3);
  --rog-border: #2a2a2a;
}

/* ROG Background Patterns */
.rog-bg {
  background: linear-gradient(135deg, var(--rog-black) 0%, var(--rog-dark) 100%);
  position: relative;
}

.rog-bg::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(circle at 20% 80%, var(--rog-accent-glow) 0%, transparent 40%),
    radial-gradient(circle at 80% 20%, rgba(255, 0, 51, 0.1) 0%, transparent 40%);
  pointer-events: none;
}

/* ROG Panel Styles */
.rog-panel {
  background: var(--rog-gray);
  border: 1px solid var(--rog-border);
  border-radius: 4px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.rog-panel-header {
  background: linear-gradient(90deg, var(--rog-gray-light) 0%, var(--rog-gray) 100%);
  border-bottom: 1px solid var(--rog-border);
  padding: 12px 16px;
}

/* ROG Button Styles */
.rog-btn {
  background: linear-gradient(180deg, var(--rog-gray-lighter) 0%, var(--rog-gray) 100%);
  border: 1px solid var(--rog-border);
  color: var(--rog-text);
  padding: 8px 16px;
  border-radius: 2px;
  transition: all 0.2s ease;
}

.rog-btn:hover {
  background: linear-gradient(180deg, var(--rog-gray-light) 0%, var(--rog-gray-lighter) 100%);
  border-color: var(--rog-accent);
  box-shadow: 0 0 10px var(--rog-accent-glow);
}

.rog-btn-primary {
  background: linear-gradient(180deg, var(--rog-red) 0%, var(--rog-red-dark) 100%);
  border-color: var(--rog-red);
  color: white;
}

.rog-btn-primary:hover {
  background: linear-gradient(180deg, var(--rog-red-dark) 0%, var(--rog-red) 100%);
  box-shadow: 0 0 20px var(--rog-accent-glow);
}

/* ROG Input Styles */
.rog-input {
  background: var(--rog-black);
  border: 1px solid var(--rog-border);
  color: var(--rog-text);
  padding: 8px 12px;
  border-radius: 2px;
}

.rog-input:focus {
  outline: none;
  border-color: var(--rog-accent);
  box-shadow: 0 0 10px var(--rog-accent-glow);
}

/* ROG Sidebar Styles */
.rog-sidebar {
  background: linear-gradient(180deg, var(--rog-gray) 0%, var(--rog-black) 100%);
  border-right: 1px solid var(--rog-border);
}

.rog-sidebar-item {
  padding: 10px 16px;
  color: var(--rog-text-dim);
  transition: all 0.2s ease;
  border-left: 3px solid transparent;
}

.rog-sidebar-item:hover {
  background: var(--rog-gray-light);
  color: var(--rog-text);
  border-left-color: var(--rog-accent);
}

.rog-sidebar-item.active {
  background: var(--rog-gray-light);
  color: var(--rog-text);
  border-left-color: var(--rog-accent);
}

/* ROG Taskbar Styles */
.rog-taskbar {
  background: linear-gradient(90deg, var(--rog-black) 0%, var(--rog-gray) 50%, var(--rog-black) 100%);
  border-top: 1px solid var(--rog-border);
}

/* ROG Logo/Brand */
.rog-brand {
  color: var(--rog-accent);
  font-weight: bold;
  text-shadow: 0 0 10px var(--rog-accent-glow);
}

/* ROG Scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--rog-black);
}

::-webkit-scrollbar-thumb {
  background: var(--rog-gray-lighter);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--rog-accent);
}

/* ROG Terminal */
.rog-terminal {
  background: var(--rog-black);
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

/* ROG Glow Effects */
.rog-glow {
  animation: rog-glow 2s ease-in-out infinite alternate;
}

@keyframes rog-glow {
  from {
    box-shadow: 0 0 5px var(--rog-accent-glow);
  }
  to {
    box-shadow: 0 0 20px var(--rog-accent-glow);
  }
}
```

- [ ] **Step 2: Update index.css to import ROG theme**

```css
@import "tailwindcss";
@import "./styles/rog-theme.css";

body {
  margin: 0;
  padding: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  background: var(--rog-black);
  color: var(--rog-text);
}
```

- [ ] **Step 3: Update Desktop.tsx with ROG classes**

```tsx
import { useState } from 'react'
import { Taskbar } from './Taskbar'
import { Sidebar } from './Sidebar'

interface App {
  id: string
  title: string
  icon: string
  component: React.ComponentType
}

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <div className="rog-terminal p-4">Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <div className="p-4">Files</div> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <div className="p-4">Monitor</div> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <div className="p-4">Agent</div> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div className="p-4">Audit</div> },
]

export function Desktop() {
  const [activeApp, setActiveApp] = useState<string | null>(null)
  
  const ActiveComponent = apps.find(a => a.id === activeApp)?.component
  
  return (
    <div className="h-screen flex flex-col rog-bg">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} activeApp={activeApp} />
        <main className="flex-1 overflow-auto p-4">
          {ActiveComponent ? (
            <ActiveComponent />
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <h1 className="text-4xl font-bold rog-brand mb-4">UNIOPS</h1>
                <p className="text-gray-400">Select an application from the sidebar</p>
              </div>
            </div>
          )}
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
```

- [ ] **Step 4: Update Taskbar.tsx with ROG classes**

```tsx
interface TaskbarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  activeApp: string | null
  onSelect: (id: string) => void
}

export function Taskbar({ apps, activeApp, onSelect }: TaskbarProps) {
  return (
    <div className="h-12 rog-taskbar flex items-center px-4">
      <div className="flex gap-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className={`rog-btn px-3 py-1 text-sm ${
              activeApp === app.id ? 'rog-btn-primary' : ''
            }`}
          >
            <span className="mr-2">{app.icon}</span>
            {app.title}
          </button>
        ))}
      </div>
      <div className="ml-auto rog-brand text-sm">
        UNIOPS
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Update Sidebar.tsx with ROG classes**

```tsx
interface SidebarProps {
  apps: Array<{ id: string; title: string; icon: string }>
  activeApp: string | null
  onSelect: (id: string) => void
}

export function Sidebar({ apps, activeApp, onSelect }: SidebarProps) {
  return (
    <div className="w-48 rog-sidebar flex flex-col">
      <div className="p-4 border-b border-gray-800">
        <h2 className="text-xl font-bold rog-brand">UNIOPS</h2>
        <p className="text-xs text-gray-500 mt-1">Agent-First Bastion</p>
      </div>
      <nav className="flex-1 p-2">
        {apps.map(app => (
          <button
            key={app.id}
            onClick={() => onSelect(app.id)}
            className={`rog-sidebar-item w-full text-left ${
              activeApp === app.id ? 'active' : ''
            }`}
          >
            <span className="mr-3">{app.icon}</span>
            {app.title}
          </button>
        ))}
      </nav>
      <div className="p-4 border-t border-gray-800">
        <div className="text-xs text-gray-500">v1.0.0</div>
      </div>
    </div>
  )
}
```

- [ ] **Step 6: Commit**

```bash
git add web/src/styles/ web/src/index.css web/src/components/
git commit -m "feat: add ROG dark theme UI"
```

---

## Task 6: GitHub Actions CI/CD

**Covers:** [S7]

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: Go project
- Produces: Automated build, test, and release

- [ ] **Step 1: Create .github/workflows/ci.yml**

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
          
      - name: Build
        run: CGO_ENABLED=0 go build -o uniops ./cmd/uniops
        
      - name: Test
        run: go test ./...
        
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
          
      - name: Build
        run: CGO_ENABLED=0 go build -o uniops.exe ./cmd/uniops
        
  build-frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: web
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '22'
          
      - name: Install dependencies
        run: npm ci
        
      - name: Build
        run: npm run build
```

- [ ] **Step 2: Create .github/workflows/release.yml**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
          
      - name: Build Linux
        run: |
          CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o uniops-linux-amd64 ./cmd/uniops
          CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o uniops-linux-arm64 ./cmd/uniops
          
      - name: Build Windows
        run: |
          CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o uniops-windows-amd64.exe ./cmd/uniops
          
      - name: Build macOS
        run: |
          CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o uniops-darwin-amd64 ./cmd/uniops
          CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o uniops-darwin-arm64 ./cmd/uniops
          
      - name: Build Frontend
        run: |
          cd web
          npm ci
          npm run build
          
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          generate_release_notes: true
          files: |
            uniops-linux-amd64
            uniops-linux-arm64
            uniops-windows-amd64.exe
            uniops-darwin-amd64
            uniops-darwin-arm64
            web/dist/*
```

- [ ] **Step 3: Commit**

```bash
mkdir -p .github/workflows
git add .github/
git commit -m "ci: add GitHub Actions for build, test, and release"
```

---

## Task 7: Docker Local Testing

**Covers:** [S7]

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.dockerignore`
- Create: `Makefile`

**Interfaces:**
- Consumes: Go project, React frontend
- Produces: Docker-based local testing environment

- [ ] **Step 1: Create Dockerfile**

```dockerfile
# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build backend
RUN CGO_ENABLED=0 go build -o uniops ./cmd/uniops

# Build frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

COPY web/ .
RUN npm run build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/uniops .
COPY --from=frontend-builder /app/web/dist ./web/dist

EXPOSE 8080

CMD ["./uniops"]
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
version: '3.8'

services:
  uniops:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
    environment:
      - JWT_SECRET=your-secret-key-here
      - DB_PATH=/root/data/uniops.db
    restart: unless-stopped
    
  # Optional: Test SSH server
  test-server:
    image: linuxserver/openssh-server
    environment:
      - PUID=1000
      - PGID=1000
      - USER_NAME=testuser
      - USER_PASSWORD=testpass
    ports:
      - "2222:2222"
```

- [ ] **Step 3: Create .dockerignore**

```
.git
.github
web/node_modules
web/dist
data
*.db
*.exe
```

- [ ] **Step 4: Create Makefile**

```makefile
.PHONY: build run test docker-build docker-run clean

# Build backend
build:
	CGO_ENABLED=0 go build -o uniops ./cmd/uniops

# Build frontend
build-frontend:
	cd web && npm run build

# Run locally
run: build
	./uniops

# Run tests
test:
	go test ./...

# Docker build
docker-build:
	docker-compose build

# Docker run
docker-run:
	docker-compose up -d

# Docker stop
docker-stop:
	docker-compose down

# Clean
clean:
	rm -f uniops uniops.exe
	rm -rf web/dist
	rm -rf data

# Development mode with hot reload
dev:
	go run ./cmd/uniops
```

- [ ] **Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml .dockerignore Makefile
git commit -m "feat: add Docker local testing environment"
```

---

## Task 8: GitHub Pages Documentation

**Covers:** [S7]

**Files:**
- Create: `docs/index.html`
- Create: `docs/styles.css`
- Create: `.github/workflows/docs.yml`

**Interfaces:**
- Consumes: None
- Produces: HTML documentation site

- [ ] **Step 1: Create docs/index.html**

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UniOps - Agent-First Bastion Host</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <div class="rog-bg min-h-screen">
        <header class="rog-panel-header">
            <div class="container mx-auto px-4">
                <h1 class="text-2xl font-bold rog-brand">UNIOPS</h1>
                <p class="text-gray-400">Agent-First Bastion Host Platform</p>
            </div>
        </header>
        
        <main class="container mx-auto px-4 py-8">
            <section class="mb-8">
                <h2 class="text-xl font-bold mb-4">Features</h2>
                <ul class="space-y-2">
                    <li class="rog-panel p-4">SSH Proxy with Auto Key Deployment</li>
                    <li class="rog-panel p-4">Agent Management (Claude, Codex, Custom)</li>
                    <li class="rog-panel p-4">Session Recording & Audit</li>
                    <li class="rog-panel p-4">Web Terminal with xterm.js</li>
                    <li class="rog-panel p-4">ROG Dark Theme UI</li>
                </ul>
            </section>
            
            <section class="mb-8">
                <h2 class="text-xl font-bold mb-4">Quick Start</h2>
                <div class="rog-panel p-4">
                    <pre class="text-sm text-gray-300">
# Docker
docker-compose up -d

# Or build from source
CGO_ENABLED=0 go build -o uniops ./cmd/uniops
./uniops
                    </pre>
                </div>
            </section>
            
            <section>
                <h2 class="text-xl font-bold mb-4">Download</h2>
                <div class="rog-panel p-4">
                    <p>Download the latest release from <a href="https://github.com/neko233/uniops/releases" class="rog-brand">GitHub Releases</a></p>
                </div>
            </section>
        </main>
        
        <footer class="rog-panel-header mt-8">
            <div class="container mx-auto px-4 text-center text-gray-500">
                <p>&copy; 2026 UniOps. Built with Go & React.</p>
            </div>
        </footer>
    </div>
</body>
</html>
```

- [ ] **Step 2: Create docs/styles.css**

```css
/* Import ROG theme */
@import url('./rog-theme.css');

/* Additional doc styles */
.container {
    max-width: 900px;
}
```

- [ ] **Step 3: Create .github/workflows/docs.yml**

```yaml
name: Deploy Docs

on:
  push:
    branches: [main]
    paths:
      - 'docs/**'

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: "pages"
  cancel-in-progress: false

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Setup Pages
        uses: actions/configure-pages@v4
        
      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: 'docs'
          
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 4: Commit**

```bash
git add docs/ .github/workflows/docs.yml
git commit -m "docs: add GitHub Pages documentation site"
```

---

## Task 9: SSH Key Deployment Handler

**Covers:** [S3, S5, S6]

**Files:**
- Create: `internal/server/handlers/sshkey.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`, `ssh.KeyManager`, `ssh.Deployer`
- Produces: API endpoints for SSH key management

- [ ] **Step 1: Create handlers/sshkey.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/ssh"
	"github.com/neko233/uniops/internal/store"
)

type SSHKeyHandler struct {
	db         *store.DB
	keyManager *ssh.KeyManager
	deployer   *ssh.Deployer
}

type GenerateKeyRequest struct {
	ServerID uint   `json:"server_id"`
	Bits     int    `json:"bits"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type DeployKeyRequest struct {
	KeyID uint `json:"key_id"`
}

func NewSSHKeyHandler(db *store.DB) *SSHKeyHandler {
	return &SSHKeyHandler{
		db:         db,
		keyManager: ssh.NewKeyManager(),
		deployer:   ssh.NewDeployer(),
	}
}

func (h *SSHKeyHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	bits := req.Bits
	if bits == 0 {
		bits = 2048
	}

	privateKey, publicKey, fingerprint, err := h.keyManager.GenerateKeyPair(bits)
	if err != nil {
		http.Error(w, "failed to generate key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sshKey := &model.SSHKey{
		Name:        server.Name + "-key",
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Fingerprint: fingerprint,
		ServerID:    req.ServerID,
		Status:      "pending",
	}

	if err := h.db.CreateSSHKey(sshKey); err != nil {
		http.Error(w, "failed to save key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sshKey)
}

func (h *SSHKeyHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	var req DeployKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	key, err := h.db.GetSSHKey(req.KeyID)
	if err != nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	server, err := h.db.GetServer(key.ServerID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	// Deploy key using password auth
	if err := h.deployer.DeployKey(server.Host, server.Port, server.Username, server.AuthData, key.PublicKey); err != nil {
		http.Error(w, "deploy failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Test key auth
	if err := h.deployer.TestKeyAuth(server.Host, server.Port, server.Username, key.PrivateKey); err != nil {
		key.Status = "deployed"
		h.db.UpdateSSHKey(key)
		http.Error(w, "key deployed but test failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	key.Status = "deployed"
	h.db.UpdateSSHKey(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "deployed",
		"key":    key,
	})
}

func (h *SSHKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	
	keys, err := h.db.GetSSHKeysByServer(uint(serverID))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *SSHKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	key, err := h.db.GetSSHKey(uint(id))
	if err != nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}
```

- [ ] **Step 2: Add store.GetSSHKey method**

```go
func (db *DB) GetSSHKey(id uint) (*model.SSHKey, error) {
	var key model.SSHKey
	err := db.First(&key, id).Error
	return &key, err
}
```

- [ ] **Step 3: Update router with SSH key routes**

```go
// SSH Key routes
sshKeyHandler := handlers.NewSSHKeyHandler(db)
r.Route("/sshkeys", func(r chi.Router) {
	r.Post("/generate", sshKeyHandler.Generate)
	r.Post("/deploy", sshKeyHandler.Deploy)
	r.Get("/server/{serverId}", sshKeyHandler.List)
	r.Get("/{id}", sshKeyHandler.Get)
})
```

- [ ] **Step 4: Commit**

```bash
git add internal/server/handlers/sshkey.go internal/server/router.go internal/store/sqlite.go
git commit -m "feat: add SSH key management API with auto-deployment"
```

---

## Task 10: Final Build & Verification

**Covers:** [S2, S7]

**Files:**
- Modify: Various

**Interfaces:**
- Consumes: All previous tasks
- Produces: Working application with all features

- [ ] **Step 1: Build Go backend (pure Go, no CGO)**

```bash
cd D:/Code/neko233-Projects/uniops
CGO_ENABLED=0 go build -o uniops ./cmd/uniops
```

- [ ] **Step 2: Build React frontend**

```bash
cd D:/Code/neko233-Projects/uniops/web
npm run build
```

- [ ] **Step 3: Docker build and test**

```bash
cd D:/Code/neko233-Projects/uniops
docker-compose build
docker-compose up -d
# Test at http://localhost:8080
docker-compose down
```

- [ ] **Step 4: Verify all features**

- Login with admin/admin
- Create a server
- Generate and deploy SSH key
- Connect via terminal
- Configure an agent

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: complete UniOps with all Phase 2 features"
```
