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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"agent":  agent.Name,
	})
}
