package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/neko233/uniops/internal/audit"
	"github.com/neko233/uniops/internal/model"
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

	sshSession, err := ssh.NewSession(
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
	defer sshSession.Close()

	// Create audit session record
	now := time.Now()
	dbSession := &model.Session{
		UserID:    1, // default; authenticated user set via middleware context if available
		ServerID:  server.ID,
		StartTime: now,
		Status:    "active",
	}
	h.db.CreateSession(dbSession)

	recorder := audit.NewRecorder(dbSession.ID)

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
				recorder.RecordInput(msg.Data)
				mu.Lock()
				sshSession.Write([]byte(msg.Data))
				mu.Unlock()
			case "resize":
				var dims struct {
					Width  int `json:"width"`
					Height int `json:"height"`
				}
				if err := json.Unmarshal([]byte(msg.Data), &dims); err == nil {
					sshSession.Resize(dims.Width, dims.Height)
				}
			}
		}
	}()

	// Read from SSH and write to WebSocket
	buf := make([]byte, 1024)
	for {
		n, err := sshSession.Read(buf)
		if err != nil {
			break
		}

		output := string(buf[:n])
		recorder.RecordOutput(output)

		msg := TerminalMessage{
			Type: "output",
			Data: output,
		}

		mu.Lock()
		conn.WriteJSON(msg)
		mu.Unlock()
	}

	// Save session replay on disconnect
	endTime := time.Now()
	dbSession.EndTime = &endTime
	dbSession.Status = "closed"
	dbSession.Replay = recorder.GetReplay()
	h.db.UpdateSession(dbSession)
}
