# UniOps Phase 4: Audit Replay, Server Selection & UI Improvements

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add audit session replay, dynamic server selection, custom agent support, and UI polish

**Architecture:** Extend existing backend with audit replay endpoint, add server selector component, enhance agent chat with custom provider, improve overall UI UX

**Tech Stack:** Go 1.26, modernc.org/sqlite, React 19, Vite, TailwindCSS

## Global Constraints

- Default port: 6020
- Pure Go SQLite (modernc.org/sqlite) - NO CGO
- ROG dark aesthetic for UI

---

## Task 1: Audit Replay API

**Covers:** [S3, S5]

**Files:**
- Create: `internal/server/handlers/audit.go`
- Modify: `internal/server/router.go`

**Interfaces:**
- Consumes: `store.DB`
- Produces: Session list, session detail, replay endpoints

- [ ] **Step 1: Create handlers/audit.go**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/store"
)

type AuditHandler struct {
	db *store.DB
}

type SessionResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	ServerID  uint   `json:"server_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
}

type ReplayEntry struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Data      string `json:"data"`
}

func NewAuditHandler(db *store.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.db.GetSessions()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var responses []SessionResponse
	for _, s := range sessions {
		resp := SessionResponse{
			ID:       s.ID,
			UserID:   s.UserID,
			ServerID: s.ServerID,
			Status:   s.Status,
		}
		if !s.StartTime.IsZero() {
			resp.StartTime = s.StartTime.Format("2006-01-02 15:04:05")
		}
		if s.EndTime != nil && !s.EndTime.IsZero() {
			resp.EndTime = s.EndTime.Format("2006-01-02 15:04:05")
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *AuditHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	session, err := h.db.GetSessionByID(uint(id))
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	resp := SessionResponse{
		ID:       session.ID,
		UserID:   session.UserID,
		ServerID: session.ServerID,
		Status:   session.Status,
	}
	if !session.StartTime.IsZero() {
		resp.StartTime = session.StartTime.Format("2006-01-02 15:04:05")
	}
	if session.EndTime != nil && !session.EndTime.IsZero() {
		resp.EndTime = session.EndTime.Format("2006-01-02 15:04:05")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuditHandler) GetReplay(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	session, err := h.db.GetSessionByID(uint(id))
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	if session.Replay == "" {
		http.Error(w, "no replay data", http.StatusNotFound)
		return
	}

	// Decode base64 replay data
	var entries []ReplayEntry
	if err := json.Unmarshal([]byte(session.Replay), &entries); err != nil {
		// Try decoding as base64 first
		import "encoding/base64"
		decoded, err := base64.StdEncoding.DecodeString(session.Replay)
		if err != nil {
			http.Error(w, "invalid replay data", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(decoded, &entries); err != nil {
			http.Error(w, "invalid replay format", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
```

- [ ] **Step 2: Update router with audit routes**

```go
// Audit routes
auditHandler := handlers.NewAuditHandler(db)
r.Route("/audit", func(r chi.Router) {
	r.Get("/sessions", auditHandler.ListSessions)
	r.Get("/sessions/{id}", auditHandler.GetSession)
	r.Get("/sessions/{id}/replay", auditHandler.GetReplay)
})
```

- [ ] **Step 3: Commit**

```bash
git add internal/server/handlers/audit.go internal/server/router.go
git commit -m "feat: add audit replay API endpoints"
```

---

## Task 2: Audit Replay UI

**Covers:** [S4]

**Files:**
- Create: `web/src/components/Audit.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: Audit API
- Produces: Session list and replay viewer

- [ ] **Step 1: Create Audit.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Window } from './Window'

interface Session {
  id: number
  user_id: number
  server_id: number
  start_time: string
  end_time: string
  status: string
}

interface ReplayEntry {
  timestamp: string
  type: string
  data: string
}

export function Audit() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [selectedSession, setSelectedSession] = useState<Session | null>(null)
  const [replay, setReplay] = useState<ReplayEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [replaying, setReplaying] = useState(false)
  const [replayIndex, setReplayIndex] = useState(0)

  useEffect(() => {
    fetchSessions()
  }, [])

  const fetchSessions = async () => {
    try {
      const res = await fetch('/api/audit/sessions')
      if (res.ok) {
        const data = await res.json()
        setSessions(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch sessions:', err)
    }
    setLoading(false)
  }

  const fetchReplay = async (sessionId: number) => {
    try {
      const res = await fetch(`/api/audit/sessions/${sessionId}/replay`)
      if (res.ok) {
        const data = await res.json()
        setReplay(data || [])
        setSelectedSession(sessions.find(s => s.id === sessionId) || null)
        setReplayIndex(0)
      }
    } catch (err) {
      console.error('Failed to fetch replay:', err)
    }
  }

  const startReplay = () => {
    if (replay.length === 0) return
    setReplaying(true)
    setReplayIndex(0)
  }

  const stopReplay = () => {
    setReplaying(false)
  }

  useEffect(() => {
    if (!replaying || replayIndex >= replay.length) {
      if (replayIndex >= replay.length && replaying) {
        setReplaying(false)
      }
      return
    }

    const timer = setTimeout(() => {
      setReplayIndex(prev => prev + 1)
    }, 100)

    return () => clearTimeout(timer)
  }, [replaying, replayIndex, replay.length])

  if (loading) {
    return <Window title="Audit"><div className="p-4 text-center">Loading...</div></Window>
  }

  return (
    <Window title="Audit Sessions">
      <div className="flex h-[calc(100vh-200px)]">
        {/* Session list */}
        <div className="w-1/3 border-r border-gray-700 overflow-auto">
          {sessions.length === 0 ? (
            <div className="p-4 text-center text-gray-500">No sessions recorded</div>
          ) : (
            <div className="space-y-2 p-2">
              {sessions.map(session => (
                <div
                  key={session.id}
                  className={`rog-panel p-3 cursor-pointer hover:bg-gray-800 ${
                    selectedSession?.id === session.id ? 'border-l-2 border-l-red-500' : ''
                  }`}
                  onClick={() => fetchReplay(session.id)}
                >
                  <div className="flex justify-between items-center">
                    <span className="font-mono text-sm">Session #{session.id}</span>
                    <span className={`text-xs px-2 py-1 rounded ${
                      session.status === 'active' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'
                    }`}>
                      {session.status}
                    </span>
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    {session.start_time}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Replay viewer */}
        <div className="flex-1 flex flex-col">
          {selectedSession ? (
            <>
              {/* Controls */}
              <div className="rog-panel-header flex items-center gap-4">
                <span className="text-sm">Session #{selectedSession.id}</span>
                <button
                  onClick={replaying ? stopReplay : startReplay}
                  className="rog-btn rog-btn-primary text-sm"
                >
                  {replaying ? 'Stop' : 'Play'}
                </button>
                <span className="text-sm text-gray-400">
                  {replayIndex} / {replay.length}
                </span>
              </div>

              {/* Terminal output */}
              <div className="flex-1 rog-terminal p-4 font-mono text-sm overflow-auto">
                {replay.slice(0, replayIndex).map((entry, i) => (
                  <div key={i} className={entry.type === 'input' ? 'text-green-400' : 'text-gray-300'}>
                    {entry.data}
                  </div>
                ))}
                {replaying && replayIndex < replay.length && (
                  <span className="animate-pulse">█</span>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-gray-500">
              Select a session to view replay
            </div>
          )}
        </div>
      </div>
    </Window>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx to use Audit**

```tsx
import { Audit } from './Audit'

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <div>Terminal</div> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={1} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <Monitor serverId={1} /> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <AgentChat agentId={1} /> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <Audit /> },
]
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/Audit.tsx web/src/components/Desktop.tsx
git commit -m "feat: add audit replay UI with session list and playback"
```

---

## Task 3: Server Selector Component

**Covers:** [S4]

**Files:**
- Create: `web/src/components/ServerSelector.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: Server API
- Produces: Server selection dropdown

- [ ] **Step 1: Create ServerSelector.tsx**

```tsx
import { useState, useEffect } from 'react'

interface Server {
  id: number
  name: string
  host: string
  status: string
}

interface ServerSelectorProps {
  onSelect: (serverId: number) => void
  selectedId: number | null
}

export function ServerSelector({ onSelect, selectedId }: ServerSelectorProps) {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchServers()
  }, [])

  const fetchServers = async () => {
    try {
      const res = await fetch('/api/servers', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      })
      if (res.ok) {
        const data = await res.json()
        setServers(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch servers:', err)
    }
    setLoading(false)
  }

  if (loading) {
    return <div className="text-gray-500 text-sm">Loading servers...</div>
  }

  if (servers.length === 0) {
    return <div className="text-gray-500 text-sm">No servers configured</div>
  }

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm text-gray-400">Server:</label>
      <select
        value={selectedId || ''}
        onChange={(e) => onSelect(Number(e.target.value))}
        className="rog-input text-sm px-2 py-1"
      >
        <option value="">Select server</option>
        {servers.map(server => (
          <option key={server.id} value={server.id}>
            {server.name} ({server.host})
          </option>
        ))}
      </select>
      {selectedId && (
        <span className={`text-xs px-2 py-1 rounded ${
          servers.find(s => s.id === selectedId)?.status === 'online'
            ? 'bg-green-900 text-green-300'
            : 'bg-gray-700 text-gray-400'
        }`}>
          {servers.find(s => s.id === selectedId)?.status || 'unknown'}
        </span>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx with server selection**

```tsx
import { useState } from 'react'
import { ServerSelector } from './ServerSelector'

export function Desktop() {
  const [activeApp, setActiveApp] = useState<string | null>(null)
  const [selectedServerId, setSelectedServerId] = useState<number | null>(null)
  
  const ActiveComponent = (() => {
    if (!activeApp) return null
    switch (activeApp) {
      case 'terminal':
        return selectedServerId ? <Terminal serverId={selectedServerId} /> : <div className="text-gray-500">Select a server first</div>
      case 'files':
        return selectedServerId ? <FileManager serverId={selectedServerId} /> : <div className="text-gray-500">Select a server first</div>
      case 'monitor':
        return selectedServerId ? <Monitor serverId={selectedServerId} /> : <div className="text-gray-500">Select a server first</div>
      case 'agent':
        return <AgentChat agentId={1} />
      case 'audit':
        return <Audit />
      default:
        return null
    }
  })()
  
  return (
    <div className="h-screen flex flex-col rog-bg">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} activeApp={activeApp} />
        <main className="flex-1 flex flex-col overflow-hidden">
          {/* Header with server selector */}
          <div className="rog-panel-header flex items-center justify-between">
            <ServerSelector onSelect={setSelectedServerId} selectedId={selectedServerId} />
          </div>
          
          {/* Content */}
          <div className="flex-1 overflow-auto p-4">
            {ActiveComponent || (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <h1 className="text-4xl font-bold rog-brand mb-4">UNIOPS</h1>
                  <p className="text-gray-400">Select an application from the sidebar</p>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ServerSelector.tsx web/src/components/Desktop.tsx
git commit -m "feat: add server selector component for dynamic server switching"
```

---

## Task 4: Custom Agent Provider

**Covers:** [S3, S5]

**Files:**
- Create: `internal/agent/custom.go`
- Modify: `internal/server/handlers/agentchat.go`

**Interfaces:**
- Consumes: `agent.Provider` interface
- Produces: Custom agent provider support

- [ ] **Step 1: Create agent/custom.go**

```go
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CustomProvider struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

type CustomRequest struct {
	Messages  []Message `json:"messages"`
	Model     string    `json:"model,omitempty"`
	Stream    bool      `json:"stream"`
}

type CustomResponse struct {
	Content string `json:"content"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewCustomProvider(apiKey, endpoint string) *CustomProvider {
	return &CustomProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

func (p *CustomProvider) Name() string {
	return "custom"
}

func (p *CustomProvider) Chat(messages []Message) (string, error) {
	req := CustomRequest{
		Messages: messages,
		Stream:   false,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try standard OpenAI-compatible format first
	var customResp CustomResponse
	if err := json.Unmarshal(respBody, &customResp); err == nil {
		if customResp.Content != "" {
			return customResp.Content, nil
		}
		if len(customResp.Choices) > 0 {
			return customResp.Choices[0].Message.Content, nil
		}
	}

	// Try plain text response
	return string(respBody), nil
}

func (p *CustomProvider) StreamChat(messages []Message) (<-chan string, error) {
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

- [ ] **Step 2: Update agentchat.go to support custom provider**

```go
// In the Chat handler, update the switch statement:
switch agentModel.Type {
case "claude":
	provider = agent.NewClaudeProvider(agentModel.APIKey, agentModel.Endpoint, "")
case "openai":
	provider = agent.NewOpenAIProvider(agentModel.APIKey, agentModel.Endpoint, "")
case "custom":
	provider = agent.NewCustomProvider(agentModel.APIKey, agentModel.Endpoint)
default:
	http.Error(w, "unsupported agent type", http.StatusBadRequest)
	return
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/agent/custom.go internal/server/handlers/agentchat.go
git commit -m "feat: add custom agent provider for OpenAI-compatible APIs"
```

---

## Task 5: Login Page

**Covers:** [S4]

**Files:**
- Create: `web/src/components/Login.tsx`
- Modify: `web/src/App.tsx`

**Interfaces:**
- Consumes: Auth API
- Produces: Login form

- [ ] **Step 1: Create Login.tsx**

```tsx
import { useState } from 'react'

interface LoginProps {
  onLogin: (token: string, user: any) => void
}

export function Login({ onLogin }: LoginProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })

      if (res.ok) {
        const data = await res.json()
        localStorage.setItem('token', data.access_token)
        onLogin(data.access_token, data.user)
      } else {
        setError('Invalid credentials')
      }
    } catch (err) {
      setError('Connection failed')
    }
    setLoading(false)
  }

  return (
    <div className="h-screen flex items-center justify-center rog-bg">
      <div className="rog-panel w-96 p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold rog-brand">UNIOPS</h1>
          <p className="text-gray-500 mt-2">Agent-First Bastion Host</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rog-input"
              required
            />
          </div>

          <div>
            <label className="block text-sm text-gray-400 mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rog-input"
              required
            />
          </div>

          {error && (
            <div className="text-red-500 text-sm">{error}</div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full rog-btn rog-btn-primary py-2"
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>

        <div className="mt-6 text-center text-xs text-gray-600">
          Default: admin / admin
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Update App.tsx with auth flow**

```tsx
import { useState, useEffect } from 'react'
import { Login } from './components/Login'
import { Desktop } from './components/Desktop'

function App() {
  const [token, setToken] = useState<string | null>(null)
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    const savedToken = localStorage.getItem('token')
    if (savedToken) {
      setToken(savedToken)
      // Verify token
      fetch('/api/auth/me', {
        headers: { 'Authorization': `Bearer ${savedToken}` },
      }).then(res => {
        if (res.ok) {
          res.json().then(data => setUser(data))
        } else {
          localStorage.removeItem('token')
          setToken(null)
        }
      })
    }
  }, [])

  const handleLogin = (newToken: string, newUser: any) => {
    setToken(newToken)
    setUser(newUser)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
  }

  if (!token) {
    return <Login onLogin={handleLogin} />
  }

  return <Desktop user={user} onLogout={handleLogout} />
}

export default App
```

- [ ] **Step 3: Update Desktop.tsx to accept user and logout props**

```tsx
interface DesktopProps {
  user: any
  onLogout: () => void
}

export function Desktop({ user, onLogout }: DesktopProps) {
  // ... existing code
  
  return (
    <div className="h-screen flex flex-col rog-bg">
      <div className="flex-1 flex overflow-hidden">
        <Sidebar apps={apps} onSelect={setActiveApp} activeApp={activeApp} />
        <main className="flex-1 flex flex-col overflow-hidden">
          <div className="rog-panel-header flex items-center justify-between">
            <ServerSelector onSelect={setSelectedServerId} selectedId={selectedServerId} />
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-400">{user?.username}</span>
              <button onClick={onLogout} className="rog-btn text-sm">Logout</button>
            </div>
          </div>
          {/* ... rest */}
        </main>
      </div>
      <Taskbar apps={apps} activeApp={activeApp} onSelect={setActiveApp} />
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/Login.tsx web/src/App.tsx web/src/components/Desktop.tsx
git commit -m "feat: add login page with auth flow"
```

---

## Task 6: Server Management UI

**Covers:** [S4]

**Files:**
- Create: `web/src/components/ServerManager.tsx`
- Modify: `web/src/components/Desktop.tsx`

**Interfaces:**
- Consumes: Server API
- Produces: Server CRUD interface

- [ ] **Step 1: Create ServerManager.tsx**

```tsx
import { useState, useEffect } from 'react'
import { Window } from './Window'

interface Server {
  id: number
  name: string
  host: string
  port: number
  username: string
  auth_type: string
  status: string
}

export function ServerManager() {
  const [servers, setServers] = useState<Server[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    host: '',
    port: 22,
    username: 'root',
    auth_type: 'password',
    auth_data: '',
  })

  useEffect(() => {
    fetchServers()
  }, [])

  const fetchServers = async () => {
    try {
      const res = await fetch('/api/servers', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
      })
      if (res.ok) {
        const data = await res.json()
        setServers(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch servers:', err)
    }
    setLoading(false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const res = await fetch('/api/servers', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(formData),
      })
      if (res.ok) {
        setShowForm(false)
        setFormData({ name: '', host: '', port: 22, username: 'root', auth_type: 'password', auth_data: '' })
        fetchServers()
      }
    } catch (err) {
      console.error('Failed to create server:', err)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this server?')) return
    try {
      await fetch(`/api/servers/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
      })
      fetchServers()
    } catch (err) {
      console.error('Failed to delete server:', err)
    }
  }

  if (loading) {
    return <Window title="Server Management"><div className="p-4 text-center">Loading...</div></Window>
  }

  return (
    <Window title="Server Management">
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-400">{servers.length} servers</span>
          <button onClick={() => setShowForm(!showForm)} className="rog-btn rog-btn-primary text-sm">
            {showForm ? 'Cancel' : 'Add Server'}
          </button>
        </div>

        {showForm && (
          <form onSubmit={handleSubmit} className="rog-panel p-4 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <input
                placeholder="Name"
                value={formData.name}
                onChange={e => setFormData({...formData, name: e.target.value})}
                className="rog-input"
                required
              />
              <input
                placeholder="Host"
                value={formData.host}
                onChange={e => setFormData({...formData, host: e.target.value})}
                className="rog-input"
                required
              />
              <input
                placeholder="Port"
                type="number"
                value={formData.port}
                onChange={e => setFormData({...formData, port: Number(e.target.value)})}
                className="rog-input"
              />
              <input
                placeholder="Username"
                value={formData.username}
                onChange={e => setFormData({...formData, username: e.target.value})}
                className="rog-input"
              />
              <select
                value={formData.auth_type}
                onChange={e => setFormData({...formData, auth_type: e.target.value})}
                className="rog-input"
              >
                <option value="password">Password</option>
                <option value="key">SSH Key</option>
              </select>
              <input
                placeholder={formData.auth_type === 'password' ? 'Password' : 'Private Key'}
                type={formData.auth_type === 'password' ? 'password' : 'text'}
                value={formData.auth_data}
                onChange={e => setFormData({...formData, auth_data: e.target.value})}
                className="rog-input"
              />
            </div>
            <button type="submit" className="rog-btn rog-btn-primary">Save</button>
          </form>
        )}

        <div className="space-y-2">
          {servers.map(server => (
            <div key={server.id} className="rog-panel p-3 flex items-center justify-between">
              <div>
                <span className="font-medium">{server.name}</span>
                <span className="text-gray-500 ml-2 text-sm">{server.host}:{server.port}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className={`text-xs px-2 py-1 rounded ${
                  server.status === 'online' ? 'bg-green-900 text-green-300' : 'bg-gray-700 text-gray-400'
                }`}>
                  {server.status || 'offline'}
                </span>
                <button onClick={() => handleDelete(server.id)} className="text-red-500 text-sm">Delete</button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </Window>
  )
}
```

- [ ] **Step 2: Update Desktop.tsx to include Server Manager**

```tsx
import { ServerManager } from './ServerManager'

const apps: App[] = [
  { id: 'terminal', title: 'Terminal', icon: '>', component: () => <Terminal serverId={selectedServerId!} /> },
  { id: 'files', title: 'Files', icon: '📁', component: () => <FileManager serverId={selectedServerId!} /> },
  { id: 'monitor', title: 'Monitor', icon: '📊', component: () => <Monitor serverId={selectedServerId!} /> },
  { id: 'agent', title: 'Agent', icon: '🤖', component: () => <AgentChat agentId={1} /> },
  { id: 'audit', title: 'Audit', icon: '📋', component: () => <Audit /> },
  { id: 'servers', title: 'Servers', icon: '🖥️', component: () => <ServerManager /> },
]
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/ServerManager.tsx web/src/components/Desktop.tsx
git commit -m "feat: add server management UI with CRUD operations"
```

---

## Task 7: Final Build & Verification

**Covers:** [S2, S7]

**Files:**
- Modify: Various

**Interfaces:**
- Consumes: All previous tasks
- Produces: Working application with all Phase 4 features

- [ ] **Step 1: Build Go backend**

```bash
cd D:/Code/neko233-Projects/uniops
$env:CGO_ENABLED="0"; go build -o uniops.exe ./cmd/uniops
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

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "feat: complete Phase 4 with audit replay, server management, and login"
```
