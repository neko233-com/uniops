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
	AgentID  uint            `json:"agent_id"`
	Messages []agent.Message `json:"messages"`
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
	case "custom":
		provider = agent.NewCustomProvider(agentModel.APIKey, agentModel.Endpoint)
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
