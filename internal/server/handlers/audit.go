package handlers

import (
	"encoding/base64"
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

	var entries []ReplayEntry
	// Try decoding as base64 first
	decoded, err := base64.StdEncoding.DecodeString(session.Replay)
	if err != nil {
		// Not base64, try raw JSON
		if err := json.Unmarshal([]byte(session.Replay), &entries); err != nil {
			http.Error(w, "invalid replay format", http.StatusInternalServerError)
			return
		}
	} else {
		if err := json.Unmarshal(decoded, &entries); err != nil {
			http.Error(w, "invalid replay data", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}
