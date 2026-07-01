# UniOps Phase 3: File Manager, Monitor, Agent UI & Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add file manager, system monitor, agent interaction UI, and Go test coverage

**Architecture:** Extend existing Go backend with file/sftp operations, system metrics collection, agent chat API, and add comprehensive tests

**Tech Stack:** Go 1.26, modernc.org/sqlite, React 19, Vite, TailwindCSS, xterm.js

## Global Constraints

- Default port: 6020
- Pure Go SQLite (modernc.org/sqlite) - NO CGO
- ROG dark aesthetic for UI
- Go tests for all backend components

---

## Task 1: Update Port to 6020

**Covers:** [S2]

**Files:**
- Modify: `cmd/uniops/main.go`
- Modify: `docker-compose.yml`
- Modify: `Dockerfile`
- Modify: `web/src/components/Terminal.tsx`

**Interfaces:**
- Consumes: None
- Produces: Port 6020 configuration

- [ ] **Step 1: Update main.go port**

```go
addr := ":6020"
```

- [ ] **Step 2: Update docker-compose.yml**

```yaml
ports:
  - "6020:6020"
```

- [ ] **Step 3: Update Dockerfile**

```dockerfile
EXPOSE 6020
```

- [ ] **Step 4: Update Terminal.tsx WebSocket**

```typescript
const ws = new WebSocket(`ws://localhost:6020/ws/terminal/${serverId}`)
```

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "config: change default port to 6020"
```

---

## Task 2: File Manager Backend

**Covers:** [S3, S5]

**Files:**
- Create: `internal/sftp/handler.go`
- Create: `internal/server/handlers/filemanager.go`

**Interfaces:**
- Consumes: `store.DB`, SSH sessions
- Produces: File list, upload, download, delete endpoints

- [ ] **Step 1: Create sftp/handler.go**

```go
package sftp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

func (h *Handler) ListFiles(host string, port int, username, password, path string) ([]FileInfo, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []FileInfo
	for _, f := range files {
		result = append(result, FileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			Mode:    f.Mode().String(),
			ModTime: f.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   f.IsDir(),
		})
	}

	return result, nil
}

func (h *Handler) DownloadFile(host string, port int, username, password, remotePath string) (io.ReadCloser, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	file, err := sftpClient.Open(remotePath)
	if err != nil {
		sftpClient.Close()
		client.Close()
		return nil, err
	}

	return file, nil
}

func (h *Handler) UploadFile(host string, port int, username, password, remotePath string, reader io.Reader) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	file, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (h *Handler) DeleteFile(host string, port int, username, password, remotePath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Remove(remotePath)
}

func (h *Handler) Mkdir(host string, port int, username, password, remotePath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Mkdir(remotePath)
}
```

- [ ] **Step 2: Create handlers/filemanager.go**

```go
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/sftp"
	"github.com/neko233/uniops/internal/store"
)

type FileManagerHandler struct {
	db      *store.DB
	handler *sftp.Handler
}

type ListFilesRequest struct {
	Path string `json:"path"`
}

type UploadRequest struct {
	Path string `json:"path"`
}

func NewFileManagerHandler(db *store.DB) *FileManagerHandler {
	return &FileManagerHandler{
		db:      db,
		handler: sftp.NewHandler(),
	}
}

func (h *FileManagerHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	var req ListFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Path = "/"
	}

	files, err := h.handler.ListFiles(server.Host, server.Port, server.Username, server.AuthData, req.Path)
	if err != nil {
		http.Error(w, "failed to list files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *FileManagerHandler) Download(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := chi.URLParam(r, "*")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	file, err := h.handler.DownloadFile(server.Host, server.Port, server.Username, server.AuthData, "/"+remotePath)
	if err != nil {
		http.Error(w, "download failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment")
	io.Copy(w, file)
}

func (h *FileManagerHandler) Upload(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := r.URL.Query().Get("path")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	if err := h.handler.UploadFile(server.Host, server.Port, server.Username, server.AuthData, remotePath, r.Body); err != nil {
		http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *FileManagerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := chi.URLParam(r, "*")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	if err := h.handler.DeleteFile(server.Host, server.Port, server.Username, server.AuthData, "/"+remotePath); err != nil {
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FileManagerHandler) Mkdir(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.handler.Mkdir(server.Host, server.Port, server.Username, server.AuthData, req.Path); err != nil {
		http.Error(w, "mkdir failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

- [ ] **Step 3: Update router with file manager routes**

```go
// File Manager routes
fileManagerHandler := handlers.NewFileManagerHandler(db)
r.Route("/files/{serverId}", func(r chi.Router) {
	r.Post("/list", fileManagerHandler.ListFiles)
	r.Get("/*", fileManagerHandler.Download)
	r.Put("/*", fileManagerHandler.Upload)
	r.Delete("/*", fileManagerHandler.Delete)
	r.Post("/mkdir", fileManagerHandler.Mkdir)
})
```

- [ ] **Step 4: Add sftp dependency**

```bash
go get github.com/pkg/sftp
```

- [ ] **Step 5: Commit**

```bash
go get github.com/pkg/sftp
git add internal/sftp/ internal/server/handlers/filemanager.go
git commit -m "feat: add file manager backend with SFTP support"
```

---

## Task 3: System Monitor Backend

**Covers:** [S3, S5]

**Files:**
- Create: `internal/monitor/collector.go`
- Create: `internal/server/handlers/monitor.go`

**Interfaces:**
- Consumes: SSH connections
- Produces: System metrics (CPU, Memory, Disk, Network)

- [ ] **Step 1: Create monitor/collector.go**

```go
package monitor

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

type SystemMetrics struct {
	CPU       CPUMetrics    `json:"cpu"`
	Memory    MemoryMetrics `json:"memory"`
	Disk      []DiskMetrics `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Hostname  string        `json:"hostname"`
	Uptime    string        `json:"uptime"`
	LoadAvg   []float64     `json:"load_avg"`
}

type CPUMetrics struct {
	Usage     float64 `json:"usage"`
	cores     int     `json:"-"`
	modelName string  `json:"-"`
}

type MemoryMetrics struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

type DiskMetrics struct {
	Mount   string  `json:"mount"`
	Size    uint64  `json:"size"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Usage   float64 `json:"usage"`
}

type NetworkMetrics struct {
	Interface string `json:"interface"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
}

func (c *Collector) Collect(host string, port int, username, password string) (*SystemMetrics, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	metrics := &SystemMetrics{}

	// Get hostname
	if output, err := c.runCommand(client, "hostname"); err == nil {
		metrics.Hostname = strings.TrimSpace(output)
	}

	// Get uptime
	if output, err := c.runCommand(client, "uptime -p"); err == nil {
		metrics.Uptime = strings.TrimSpace(output)
	}

	// Get load average
	if output, err := c.runCommand(client, "cat /proc/loadavg"); err == nil {
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			fmt.Sscanf(parts[0], "%f", &metrics.LoadAvg[0])
			fmt.Sscanf(parts[1], "%f", &metrics.LoadAvg[1])
			fmt.Sscanf(parts[2], "%f", &metrics.LoadAvg[2])
		}
	}

	// Get CPU usage
	if output, err := c.runCommand(client, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}'"); err == nil {
		var usage float64
		fmt.Sscanf(strings.TrimSpace(output), "%f", &usage)
		metrics.CPU.Usage = usage
	}

	// Get memory info
	if output, err := c.runCommand(client, "free -b | grep Mem"); err == nil {
		parts := strings.Fields(output)
		if len(parts) >= 4 {
			fmt.Sscanf(parts[1], "%d", &metrics.Memory.Total)
			fmt.Sscanf(parts[2], "%d", &metrics.Memory.Used)
			fmt.Sscanf(parts[3], "%d", &metrics.Memory.Free)
			metrics.Memory.Usage = float64(metrics.Memory.Used) / float64(metrics.Memory.Total) * 100
		}
	}

	// Get disk info
	if output, err := c.runCommand(client, "df -B1 | grep -E '^/dev/'"); err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				var disk DiskMetrics
				disk.Mount = parts[5]
				fmt.Sscanf(parts[1], "%d", &disk.Size)
				fmt.Sscanf(parts[2], "%d", &disk.Used)
				fmt.Sscanf(parts[3], "%d", &disk.Free)
				disk.Usage = float64(disk.Used) / float64(disk.Size) * 100
				metrics.Disk = append(metrics.Disk, disk)
			}
		}
	}

	// Get network info
	if output, err := c.runCommand(client, "cat /proc/net/dev | grep -v 'face'"); err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 10 {
				iface := strings.Trim(parts[0], ":")
				if iface == "lo" {
					continue
				}
				var net NetworkMetrics
				net.Interface = iface
				fmt.Sscanf(parts[1], "%d", &net.RxBytes)
				fmt.Sscanf(parts[9], "%d", &net.TxBytes)
				metrics.Network = append(metrics.Network, net)
			}
		}
	}

	return metrics, nil
}

func (c *Collector) runCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}
```

- [ ] **Step 2: Create handlers/monitor.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/monitor"
	"github.com/neko233/uniops/internal/store"
)

type MonitorHandler struct {
	db        *store.DB
	collector *monitor.Collector
}

func NewMonitorHandler(db *store.DB) *MonitorHandler {
	return &MonitorHandler{
		db:        db,
		collector: monitor.NewCollector(),
	}
}

func (h *MonitorHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	if err != nil {
		http.Error(w, "invalid server id", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	metrics, err := h.collector.Collect(server.Host, server.Port, server.Username, server.AuthData)
	if err != nil {
		http.Error(w, "failed to collect metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
```

- [ ] **Step 3: Update router with monitor routes**

```go
// Monitor routes
monitorHandler := handlers.NewMonitorHandler(db)
r.Get("/monitor/{serverId}", monitorHandler.GetMetrics)
```

- [ ] **Step 4: Commit**

```bash
git add internal/monitor/ internal/server/handlers/monitor.go
git commit -m "feat: add system monitor backend with CPU, memory, disk, network metrics"
```

---

## Task 4: Agent Chat API

**Covers:** [S3, S5]

**Files:**
- Create: `internal/agent/provider.go`
- Create: `internal/agent/claude.go`
- Create: `internal/agent/openai.go`
- Create: `internal/server/handlers/agentchat.go`

**Interfaces:**
- Consumes: `store.DB`, Agent configs
- Produces: Agent chat/completion endpoints

- [ ] **Step 1: Create agent/provider.go**

```go
package agent

type Provider interface {
	Name() string
	Chat(messages []Message) (string, error)
	StreamChat(messages []Message) (<-chan string, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Content string `json:"content"`
}
```

- [ ] **Step 2: Create agent/claude.go**

```go
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClaudeProvider struct {
	apiKey     string
	endpoint   string
	model      string
	httpClient *http.Client
}

type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type ClaudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewClaudeProvider(apiKey, endpoint, model string) *ClaudeProvider {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &ClaudeProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{},
	}
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) Chat(messages []Message) (string, error) {
	req := ClaudeRequest{
		Model:     p.model,
		MaxTokens: 4096,
		Messages:  messages,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", err
	}

	if len(claudeResp.Content) > 0 {
		return claudeResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}

func (p *ClaudeProvider) StreamChat(messages []Message) (<-chan string, error) {
	// Simplified: non-streaming fallback
	ch := make(chan string, 1)
	result, err := p.Chat(messages)
	if err != nil {
		close(ch)
		return nil, err
	}
	ch <- result
	close(ch)
	return ch, nil
}
```

- [ ] **Step 3: Create agent/openai.go**

```go
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAIProvider struct {
	apiKey     string
	endpoint   string
	model      string
	httpClient *http.Client
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewOpenAIProvider(apiKey, endpoint, model string) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4"
	}
	return &OpenAIProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Chat(messages []Message) (string, error) {
	req := OpenAIRequest{
		Model:    p.model,
		Messages: messages,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return "", err
	}

	if len(openaiResp.Choices) > 0 {
		return openaiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content in response")
}

func (p *OpenAIProvider) StreamChat(messages []Message) (<-chan string, error) {
	ch := make(chan string, 1)
	result, err := p.Chat(messages)
	if err != nil {
		close(ch)
		return nil, err
	}
	ch <- result
	close(ch)
	return ch, nil
}
```

- [ ] **Step 4: Create handlers/agentchat.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neko233/uniops/internal/agent"
	"github.com/neko233/uniops/internal/store"
)

type AgentChatHandler struct {
	db *store.DB
}

type ChatRequest struct {
	AgentID  uint             `json:"agent_id"`
	Messages []agent.Message  `json:"messages"`
}

func NewAgentChatHandler(db *store.DB) *AgentChatHandler {
	return &AgentChatHandler{db: db}
}

func (h *AgentChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agentModel, err := h.db.GetAgent(req.AgentID)
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	var provider agent.Provider
	switch agentModel.Type {
	case "claude":
		provider = agent.NewClaudeProvider(agentModel.APIKey, agentModel.Endpoint, "")
	case "openai":
		provider = agent.NewOpenAIProvider(agentModel.APIKey, agentModel.Endpoint, "")
	default:
		http.Error(w, "unsupported agent type", http.StatusBadRequest)
		return
	}

	response, err := provider.Chat(req.Messages)
	if err != nil {
		http.Error(w, "agent error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent.ChatResponse{Content: response})
}
```

- [ ] **Step 5: Update router with agent chat route**

```go
// Agent Chat routes
agentChatHandler := handlers.NewAgentChatHandler(db)
r.Post("/agent/chat", agentChatHandler.Chat)
```

- [ ] **Step 6: Commit**

```bash
git add internal/agent/ internal/server/handlers/agentchat.go
git commit -m "feat: add agent chat API with Claude and OpenAI providers"
```

---

## Task 5: File Manager UI

**Covers:** [S4]

**Files:**
- Create: `web/src/components/FileManager.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: File manager API
- Produces: File browser UI

- [ ] **Step 1: Create FileManager.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Window } from './Window'

interface FileItem {
  name: string
  size: number
  mode: string
  mod_time: string
  is_dir: boolean
}

interface FileManagerProps {
  serverId: number
}

export function FileManager({ serverId }: FileManagerProps) {
  const [files, setFiles] = useState<FileItem[]>([])
  const [currentPath, setCurrentPath] = useState('/')
  const [loading, setLoading] = useState(false)

  const fetchFiles = async (path: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/files/${serverId}/list`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      })
      if (res.ok) {
        const data = await res.json()
        setFiles(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch files:', err)
    }
    setLoading(false)
  }

  useEffect(() => {
    fetchFiles(currentPath)
  }, [currentPath])

  const handleDoubleClick = (file: FileItem) => {
    if (file.is_dir) {
      setCurrentPath(`${currentPath === '/' ? '' : currentPath}/${file.name}`)
    }
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  return (
    <Window title="File Manager">
      <div className="h-full flex flex-col">
        {/* Path bar */}
        <div className="rog-panel-header flex items-center gap-2 mb-2">
          <button 
            className="rog-btn px-2 py-1"
            onClick={() => setCurrentPath('/')}
          >
            /
          </button>
          {currentPath.split('/').filter(Boolean).map((part, i, arr) => (
            <span key={i} className="flex items-center">
              <span className="text-gray-500">/</span>
              <button
                className="rog-btn px-2 py-1"
                onClick={() => setCurrentPath('/' + arr.slice(0, i + 1).join('/'))}
              >
                {part}
              </button>
            </span>
          ))}
        </div>

        {/* File list */}
        <div className="flex-1 overflow-auto rog-panel">
          {loading ? (
            <div className="p-4 text-center text-gray-400">Loading...</div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="text-left p-2">Name</th>
                  <th className="text-right p-2">Size</th>
                  <th className="text-right p-2">Modified</th>
                  <th className="text-right p-2">Mode</th>
                </tr>
              </thead>
              <tbody>
                {files.map((file) => (
                  <tr
                    key={file.name}
                    className="border-b border-gray-800 hover:bg-gray-800 cursor-pointer"
                    onDoubleClick={() => handleDoubleClick(file)}
                  >
                    <td className="p-2">
                      <span className="mr-2">{file.is_dir ? '📁' : '📄'}</span>
                      {file.name}
                    </td>
                    <td className="p-2 text-right text-gray-400">
                      {file.is_dir ? '-' : formatSize(file.size)}
                    </td>
                    <td className="p-2 text-right text-gray-400">
                      {file.mod_time}
                    </td>
                    <td className="p-2 text-right text-gray-400 font-mono text-sm">
                      {file.mode}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </Window>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx to use FileManager**

```tsx
import { FileManager } from './FileManager'

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <div>Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={1} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <div>Monitor</div> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <div>Agent</div> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div>Audit</div> },
]
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/FileManager.tsx web/src/components/Desktop.tsx
git commit -m "feat: add file manager UI component"
```

---

## Task 6: System Monitor UI

**Covers:** [S4]

**Files:**
- Create: `web/src/components/Monitor.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: Monitor API
- Produces: Real-time system metrics display

- [ ] **Step 1: Create Monitor.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Window } from './Window'

interface SystemMetrics {
  cpu: { usage: number }
  memory: { total: number; used: number; usage: number }
  disk: Array<{ mount: string; size: number; used: number; usage: number }>
  hostname: string
  uptime: string
  load_avg: number[]
}

interface MonitorProps {
  serverId: number
}

export function Monitor({ serverId }: MonitorProps) {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchMetrics = async () => {
    try {
      const res = await fetch(`/api/monitor/${serverId}`)
      if (res.ok) {
        const data = await res.json()
        setMetrics(data)
      }
    } catch (err) {
      console.error('Failed to fetch metrics:', err)
    }
    setLoading(false)
  }

  useEffect(() => {
    fetchMetrics()
    const interval = setInterval(fetchMetrics, 5000)
    return () => clearInterval(interval)
  }, [serverId])

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }

  const getUsageColor = (usage: number) => {
    if (usage < 50) return 'bg-green-500'
    if (usage < 80) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  if (loading) {
    return <Window title="System Monitor"><div className="p-4 text-center">Loading...</div></Window>
  }

  if (!metrics) {
    return <Window title="System Monitor"><div className="p-4 text-center">Failed to load metrics</div></Window>
  }

  return (
    <Window title="System Monitor">
      <div className="space-y-4">
        {/* Header */}
        <div className="rog-panel p-4">
          <div className="flex justify-between items-center">
            <div>
              <h3 className="text-lg font-bold">{metrics.hostname}</h3>
              <p className="text-gray-400 text-sm">{metrics.uptime}</p>
            </div>
            <div className="text-right">
              <p className="text-gray-400">Load Average</p>
              <p className="font-mono">{metrics.load_avg?.map(l => l.toFixed(2)).join(' ')}</p>
            </div>
          </div>
        </div>

        {/* CPU */}
        <div className="rog-panel p-4">
          <h4 className="mb-2">CPU</h4>
          <div className="flex items-center gap-4">
            <div className="flex-1 rog-input h-6 rounded overflow-hidden">
              <div 
                className={`h-full ${getUsageColor(metrics.cpu.usage)}`}
                style={{ width: `${metrics.cpu.usage}%` }}
              />
            </div>
            <span className="w-16 text-right">{metrics.cpu.usage.toFixed(1)}%</span>
          </div>
        </div>

        {/* Memory */}
        <div className="rog-panel p-4">
          <h4 className="mb-2">Memory</h4>
          <div className="flex items-center gap-4">
            <div className="flex-1 rog-input h-6 rounded overflow-hidden">
              <div 
                className={`h-full ${getUsageColor(metrics.memory.usage)}`}
                style={{ width: `${metrics.memory.usage}%` }}
              />
            </div>
            <span className="w-32 text-right text-sm">
              {formatBytes(metrics.memory.used)} / {formatBytes(metrics.memory.total)}
            </span>
          </div>
        </div>

        {/* Disk */}
        <div className="rog-panel p-4">
          <h4 className="mb-2">Disk</h4>
          <div className="space-y-2">
            {metrics.disk?.map((d) => (
              <div key={d.mount} className="flex items-center gap-4">
                <span className="w-24 text-sm truncate">{d.mount}</span>
                <div className="flex-1 rog-input h-4 rounded overflow-hidden">
                  <div 
                    className={`h-full ${getUsageColor(d.usage)}`}
                    style={{ width: `${d.usage}%` }}
                  />
                </div>
                <span className="w-32 text-right text-sm">
                  {formatBytes(d.used)} / {formatBytes(d.size)}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Network */}
        <div className="rog-panel p-4">
          <h4 className="mb-2">Network</h4>
          <div className="space-y-2">
            {metrics.network?.map((n) => (
              <div key={n.interface} className="flex items-center gap-4">
                <span className="w-24 text-sm">{n.interface}</span>
                <span className="text-sm text-green-400">↓ {formatBytes(n.rx_bytes)}</span>
                <span className="text-sm text-blue-400">↑ {formatBytes(n.tx_bytes)}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </Window>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx to use Monitor**

```tsx
import { Monitor } from './Monitor'

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <div>Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={1} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <Monitor serverId={1} /> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <div>Agent</div> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div>Audit</div> },
]
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/Monitor.tsx web/src/components/Desktop.tsx
git commit -m "feat: add system monitor UI with real-time metrics"
```

---

## Task 7: Agent Chat UI

**Covers:** [S4]

**Files:**
- Create: `web/src/components/AgentChat.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: Agent chat API
- Produces: Chat interface for AI agents

- [ ] **Step 1: Create AgentChat.tsx**

```tsx
import { useState } from 'react'
import { Window } from './Window'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

interface AgentChatProps {
  agentId: number
}

export function AgentChat({ agentId }: AgentChatProps) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)

  const sendMessage = async () => {
    if (!input.trim() || loading) return

    const userMessage: Message = { role: 'user', content: input }
    setMessages(prev => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch('/api/agent/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          agent_id: agentId,
          messages: [...messages, userMessage].map(m => ({
            role: m.role,
            content: m.content,
          })),
        }),
      })

      if (res.ok) {
        const data = await res.json()
        setMessages(prev => [...prev, { role: 'assistant', content: data.content }])
      }
    } catch (err) {
      console.error('Chat error:', err)
    }
    setLoading(false)
  }

  return (
    <Window title="Agent Chat">
      <div className="flex flex-col h-[calc(100vh-200px)]">
        {/* Messages */}
        <div className="flex-1 overflow-auto space-y-4 mb-4">
          {messages.length === 0 && (
            <div className="text-center text-gray-500 py-8">
              Start a conversation with the AI agent
            </div>
          )}
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`rog-panel p-3 ${
                msg.role === 'user' ? 'ml-12' : 'mr-12'
              }`}
            >
              <div className="text-xs text-gray-500 mb-1">
                {msg.role === 'user' ? 'You' : 'Agent'}
              </div>
              <div className="whitespace-pre-wrap">{msg.content}</div>
            </div>
          ))}
          {loading && (
            <div className="rog-panel p-3 mr-12">
              <div className="text-xs text-gray-500 mb-1">Agent</div>
              <div className="text-gray-400">Thinking...</div>
            </div>
          )}
        </div>

        {/* Input */}
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
            placeholder="Type a message..."
            className="flex-1 rog-input"
            disabled={loading}
          />
          <button
            onClick={sendMessage}
            disabled={loading || !input.trim()}
            className="rog-btn rog-btn-primary"
          >
            Send
          </button>
        </div>
      </div>
    </Window>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx to use AgentChat**

```tsx
import { AgentChat } from './AgentChat'

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <div>Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={1} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <Monitor serverId={1} /> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <AgentChat agentId={1} /> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <div>Audit</div> },
]
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/AgentChat.tsx web/src/components/Desktop.tsx
git commit -m "feat: add agent chat UI with message history"
```

---

## Task 8: Go Tests

**Covers:** [S7]

**Files:**
- Create: `internal/store/sqlite_test.go`
- Create: `internal/auth/jwt_test.go`
- Create: `internal/ssh/keymanager_test.go`

**Interfaces:**
- Consumes: Existing backend code
- Produces: Test coverage

- [ ] **Step 1: Create store/sqlite_test.go**

```go
package store

import (
	"os"
	"testing"

	"github.com/neko233/uniops/internal/model"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	
	user := &model.User{
		Username: "testuser",
		Password: "hashedpassword",
		Role:     "operator",
	}
	
	if err := db.CreateUser(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	got, err := db.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	
	if got.Username != user.Username {
		t.Errorf("Username = %v, want %v", got.Username, user.Username)
	}
	if got.Role != user.Role {
		t.Errorf("Role = %v, want %v", got.Role, user.Role)
	}
}

func TestCreateAndGetServer(t *testing.T) {
	db := setupTestDB(t)
	
	server := &model.Server{
		Name:     "Test Server",
		Host:     "192.168.1.100",
		Port:     22,
		Username: "root",
		AuthType: "password",
		AuthData: "secret",
	}
	
	if err := db.CreateServer(server); err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	
	got, err := db.GetServer(server.ID)
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}
	
	if got.Name != server.Name {
		t.Errorf("Name = %v, want %v", got.Name, server.Name)
	}
}

func TestListServers(t *testing.T) {
	db := setupTestDB(t)
	
	for i := 0; i < 3; i++ {
		server := &model.Server{
			Name: "Server",
			Host: "192.168.1.100",
			Port: 22,
		}
		db.CreateServer(server)
	}
	
	servers, err := db.GetServers()
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}
	
	if len(servers) != 3 {
		t.Errorf("Got %d servers, want 3", len(servers))
	}
}

func TestDeleteServer(t *testing.T) {
	db := setupTestDB(t)
	
	server := &model.Server{
		Name: "To Delete",
		Host: "192.168.1.100",
		Port: 22,
	}
	db.CreateServer(server)
	
	if err := db.DeleteServer(server.ID); err != nil {
		t.Fatalf("Failed to delete server: %v", err)
	}
	
	_, err := db.GetServer(server.ID)
	if err == nil {
		t.Error("Expected error getting deleted server")
	}
}

func TestInitAdmin(t *testing.T) {
	db := setupTestDB(t)
	
	if err := db.InitAdmin(); err != nil {
		t.Fatalf("Failed to init admin: %v", err)
	}
	
	admin, err := db.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("Failed to get admin: %v", err)
	}
	
	if admin.Role != "admin" {
		t.Errorf("Admin role = %v, want admin", admin.Role)
	}
}
```

- [ ] **Step 2: Create auth/jwt_test.go**

```go
package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret")
	
	token, err := manager.GenerateAccessToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if claims.UserID != 1 {
		t.Errorf("UserID = %d, want 1", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("Username = %v, want admin", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %v, want admin", claims.Role)
	}
}

func TestInvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")
	
	_, err := manager.ValidateToken("invalid-token")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestWrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1")
	manager2 := NewJWTManager("secret-2")
	
	token, _ := manager1.GenerateAccessToken(1, "admin", "admin")
	
	_, err := manager2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken with wrong secret, got %v", err)
	}
}
```

- [ ] **Step 3: Create ssh/keymanager_test.go**

```go
package ssh

import (
	"strings"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	km := NewKeyManager()
	
	privateKey, publicKey, fingerprint, err := km.GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	if privateKey == "" {
		t.Error("Private key is empty")
	}
	if !strings.HasPrefix(publicKey, "ssh-rsa") {
		t.Errorf("Public key doesn't start with ssh-rsa: %v", publicKey)
	}
	if fingerprint == "" {
		t.Error("Fingerprint is empty")
	}
}

func TestGetFingerprint(t *testing.T) {
	km := NewKeyManager()
	
	privateKey, _, fingerprint1, _ := km.GenerateKeyPair(2048)
	
	fingerprint2, err := km.GetFingerprint(privateKey)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	if fingerprint1 != fingerprint2 {
		t.Errorf("Fingerprints don't match: %v != %v", fingerprint1, fingerprint2)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd D:/Code/neko233-Projects/uniops
go test ./internal/store/ ./internal/auth/ ./internal/ssh/ -v
```

- [ ] **Step 5: Commit**

```bash
git add internal/store/sqlite_test.go internal/auth/jwt_test.go internal/ssh/keymanager_test.go
git commit -m "test: add Go unit tests for store, auth, and SSH key manager"
```

---

## Task 9: Final Build & Verification

**Covers:** [S2, S7]

**Files:**
- Modify: Various

**Interfaces:**
- Consumes: All previous tasks
- Produces: Working application with all features

- [ ] **Step 1: Build Go backend**

```bash
cd D:/Code/neko233-Projects/uniops
CGO_ENABLED=0 go build -o uniops.exe ./cmd/uniops
```

- [ ] **Step 2: Build React frontend**

```bash
cd D:/Code/neko233-Projects/uniops/web
npm run build
```

- [ ] **Step 3: Run all tests**

```bash
go test ./... -v
```

- [ ] **Step 4: Docker build**

```bash
docker-compose build
```

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: complete Phase 3 with file manager, monitor, agent UI, and tests"
```
