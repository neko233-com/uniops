# UniOps - Agent-First Bastion Host Platform

## [S1] Problem

运维团队需要一个集中化的堡垒机平台，能够：
- 管理多台服务器的 SSH 连接
- 集成 AI Agent（Claude、Codex）辅助运维操作
- 提供可视化操作界面，类似 GMSSH 的桌面体验
- 记录所有操作用于审计

## [S2] Solution Overview

UniOps 是一个 Agent-First 的堡垒机运维平台，采用 Go 后端 + SQLite + React 前端架构。

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                     Web Browser (Desktop UI)                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │ Terminal  │ │ File Mgr │ │ Monitor  │ │ Agent    │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │ WebSocket / HTTP
┌─────────────────────────────────────────────────────────────┐
│                    Go Backend (UniOps Server)                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │ SSH Proxy│ │ Agent Mgr│ │ Audit Log│ │ RBAC     │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
└─────────────────────────────────────────────────────────────┘
                           │ gRPC / HTTP
              ┌────────────┴────────────┐
              │                         │
    ┌─────────▼─────────┐     ┌─────────▼─────────┐
    │  Target Server    │     │  Target Server    │
    │  (with Agent)     │     │  (with Agent)     │
    └───────────────────┘     └───────────────────┘
```

### 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26 + Chi Router + WebSocket |
| 数据库 | SQLite + GORM |
| 前端 | React 19 + Vite + TypeScript |
| UI 组件 | Radix UI + TailwindCSS |
| 终端 | xterm.js |
| Agent | Claude API / OpenAI API / Custom |

## [S3] Architecture Design

### 3.1 Backend Structure

```
uniops/
├── cmd/
│   └── uniops/
│       └── main.go
├── internal/
│   ├── server/          # HTTP/WS server
│   │   ├── router.go
│   │   ├── middleware.go
│   │   └── handlers/
│   ├── ssh/             # SSH proxy
│   │   ├── proxy.go
│   │   ├── session.go
│   │   └── recorder.go
│   ├── agent/           # Agent management
│   │   ├── manager.go
│   │   ├── claude.go
│   │   ├── openai.go
│   │   └── custom.go
│   ├── auth/            # Authentication & RBAC
│   │   ├── jwt.go
│   │   ├── rbac.go
│   │   └── middleware.go
│   ├── audit/           # Audit logging
│   │   ├── recorder.go
│   │   └── replay.go
│   ├── model/           # Data models
│   │   ├── user.go
│   │   ├── server.go
│   │   ├── session.go
│   │   └── agent.go
│   └── store/           # Database layer
│       └── sqlite.go
├── web/                 # Frontend (React)
│   ├── src/
│   ├── index.html
│   └── vite.config.ts
├── go.mod
└── go.sum
```

### 3.2 Data Models

```go
// User model
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex"`
    Password  string
    Role      string    // admin, operator, viewer
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Server model
type Server struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    Host      string
    Port      int       `default:22`
    Username  string
    AuthType  string    // password, key
    AuthData  string    // encrypted
    AgentID   uint      // linked agent
    GroupID   uint
    Status    string    // online, offline
    CreatedAt time.Time
}

// Agent model
type Agent struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    Type      string    // claude, openai, custom
    Endpoint  string    // API endpoint
    APIKey    string    // encrypted
    Config    string    // JSON config
    Status    string    // active, inactive
    CreatedAt time.Time
}

// Session model (audit)
type Session struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint
    ServerID  uint
    StartTime time.Time
    EndTime   time.Time
    Status    string    // active, closed
    Replay    string    // base64 encoded replay data
}

// Command model (audit)
type Command struct {
    ID        uint      `gorm:"primaryKey"`
    SessionID uint
    Command   string
    Output    string
    Timestamp time.Time
}
```

### 3.3 API Endpoints

```
POST   /api/auth/login
POST   /api/auth/logout
GET    /api/auth/me

GET    /api/users
POST   /api/users
PUT    /api/users/:id
DELETE /api/users/:id

GET    /api/servers
POST   /api/servers
PUT    /api/servers/:id
DELETE /api/servers/:id
POST   /api/servers/:id/test

GET    /api/agents
POST   /api/agents
PUT    /api/agents/:id
DELETE /api/agents/:id
POST   /api/agents/:id/test

GET    /api/sessions
GET    /api/sessions/:id
GET    /api/sessions/:id/replay

WS     /ws/terminal/:serverId
WS     /ws/agent/:agentId
```

## [S4] Frontend Design

### 4.1 Desktop UI Layout

```
┌─────────────────────────────────────────────────────────────┐
│  UniOps                          [User] [Settings] [Logout] │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────┐ ┌─────────────────────────────────────────┐   │
│  │ Apps    │ │                                         │   │
│  │ ─────── │ │            Active Window                │   │
│  │ 🖥 Term │ │                                         │   │
│  │ 📁 Files│ │                                         │   │
│  │ 📊 Mon  │ │                                         │   │
│  │ 🤖 Agent│ │                                         │   │
│  │ 📋 Audit│ │                                         │   │
│  │         │ │                                         │   │
│  │ Servers │ │                                         │   │
│  │ ─────── │ │                                         │   │
│  │ • Web-1 │ │                                         │   │
│  │ • DB-1  │ │                                         │   │
│  │ • App-1 │ │                                         │   │
│  └─────────┘ └─────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  [Terminal] [Files] [Monitor] [Agent] [Audit]    CPU: 23%  │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Key Pages

1. **Terminal** - xterm.js based SSH terminal with multiple tabs
2. **File Manager** - SFTP-style file browser with upload/download
3. **Monitor** - Real-time system metrics (CPU, Memory, Disk, Network)
4. **Agent** - Configure and interact with AI agents
5. **Audit** - Session replay and command history search
6. **Settings** - User management, server management, agent configuration

## [S5] Implementation Plan

### Phase 1: Core Foundation
1. Initialize Go project with dependencies
2. Implement SQLite store layer
3. Create auth system (JWT + RBAC)
4. Build HTTP router with middleware
5. Implement WebSocket handler

### Phase 2: SSH & Terminal
1. SSH proxy implementation
2. Terminal WebSocket bridge
3. Session recording

### Phase 3: Agent Integration
1. Agent manager
2. Claude API integration
3. OpenAI API integration
4. Custom agent support

### Phase 4: Frontend
1. React project setup
2. Desktop UI layout
3. Terminal component
4. File manager
5. Monitor dashboard
6. Agent interface
7. Audit viewer

### Phase 5: Polish
1. Error handling
2. Security hardening
3. Documentation
4. Deployment scripts

## [S6] Security Considerations

- All passwords/API keys encrypted at rest
- JWT with short expiry, refresh tokens
- RBAC: admin (full access), operator (SSH + agent), viewer (read-only)
- All commands recorded for audit
- WebSocket connections authenticated
- Rate limiting on API endpoints
- Input sanitization

## [S7] Success Criteria

1. Can connect to target servers via Web Terminal
2. Can manage servers, users, agents via UI
3. Can record and replay sessions
4. Can configure and use Claude/Codex agents
5. Desktop-like UI experience in browser
6. All operations audited
