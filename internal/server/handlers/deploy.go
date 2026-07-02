package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/deploy"
	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/store"
)

type DeployHandler struct {
	db      *store.DB
	deployer *deploy.Service
}

type DeployRequest struct {
	ServerID    uint   `json:"server_id"`
	Type        string `json:"type"` // nginx, backend, full
	ServiceName string `json:"service_name"`
	BinaryURL   string `json:"binary_url"`
	AppPort     int    `json:"app_port"`
	Domain      string `json:"domain"`
	NginxPort   int    `json:"nginx_port"`
}

func NewDeployHandler(db *store.DB) *DeployHandler {
	return &DeployHandler{
		db:       db,
		deployer: deploy.NewService(db),
	}
}

func (h *DeployHandler) List(w http.ResponseWriter, r *http.Request) {
	deployments, err := h.db.GetDeployments()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func (h *DeployHandler) ListByServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	if err != nil {
		http.Error(w, "invalid server id", http.StatusBadRequest)
		return
	}
	deployments, err := h.db.GetDeploymentsByServer(uint(id))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func (h *DeployHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	deployment, err := h.db.GetDeployment(uint(id))
	if err != nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployment)
}

func (h *DeployHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ServerID == 0 {
		http.Error(w, "server_id required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "full"
	}

	cfg := deploy.Config{
		ServiceName: req.ServiceName,
		BinaryURL:   req.BinaryURL,
		AppPort:     req.AppPort,
		Domain:      req.Domain,
		NginxPort:   req.NginxPort,
	}
	cfgJSON, _ := json.Marshal(cfg)

	deployment := &model.Deployment{
		ServerID:    req.ServerID,
		Type:        req.Type,
		Config:      string(cfgJSON),
		TriggeredBy: "manual",
		Status:      "pending",
	}

	if err := h.db.CreateDeployment(deployment); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Run deployment async
	go func() {
		if err := h.deployer.Deploy(deployment.ID, func(s string) {
			log.Println(s)
		}); err != nil {
			log.Printf("Deploy %d failed: %v", deployment.ID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

// Exec runs an arbitrary command on a server (for agent integration)
func (h *DeployHandler) Exec(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerID uint   `json:"server_id"`
		Command  string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	output, err := h.deployer.ExecCommand(server, req.Command)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"output": output,
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"output": output,
	})
}

// Watch streams deployment logs via WebSocket
func (h *DeployHandler) Watch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deployment, err := h.db.GetDeployment(uint(id))
	if err != nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	var mu sync.Mutex

	sendLog := func(msg string) {
		mu.Lock()
		conn.WriteJSON(map[string]string{"type": "log", "data": msg})
		mu.Unlock()
	}

	sendLog(fmt.Sprintf("Watching deployment #%d (%s)...\n", deployment.ID, deployment.Type))

	if deployment.Status == "completed" || deployment.Status == "failed" {
		sendLog(deployment.Logs)
		sendLog(fmt.Sprintf("\nDeployment status: %s", deployment.Status))
		return
	}

	// Re-run and stream
	go func() {
		err := h.deployer.Deploy(deployment.ID, sendLog)
		if err != nil {
			sendLog(fmt.Sprintf("\n[error] %v", err))
		}
		mu.Lock()
		conn.WriteJSON(map[string]string{"type": "done", "data": "deployment finished"})
		mu.Unlock()
	}()

	// Keep connection alive until client disconnects
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
