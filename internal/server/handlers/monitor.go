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
