package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neko233/uniops/internal/oplog"
)

type OpLogHandler struct {
	logger *oplog.Logger
}

func NewOpLogHandler(logger *oplog.Logger) *OpLogHandler {
	return &OpLogHandler{logger: logger}
}

// ListDates returns available log dates
func (h *OpLogHandler) ListDates(w http.ResponseWriter, r *http.Request) {
	dates, err := h.logger.ListDates()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if dates == nil {
		dates = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dates)
}

// ReadDate returns logs for a specific date
func (h *OpLogHandler) ReadDate(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}

	entries, err := h.logger.ReadDate(date)
	if err != nil {
		http.Error(w, "no logs for this date", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// Search logs across all dates
func (h *OpLogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	user := q.Get("user")
	action := q.Get("action")
	keyword := q.Get("keyword")

	entries, err := h.logger.Search(user, action, keyword, 500)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []oplog.Entry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// DeleteRange deletes logs in a date range
func (h *OpLogHandler) DeleteRange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.StartDate == "" || req.EndDate == "" {
		http.Error(w, "start_date and end_date required", http.StatusBadRequest)
		return
	}

	deleted, err := h.logger.DeleteRange(req.StartDate, req.EndDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": deleted,
	})
}
