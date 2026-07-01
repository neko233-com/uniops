package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/store"
)

type ServerHandler struct {
	db *store.DB
}

type CreateServerRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	AuthType string `json:"auth_type"`
	AuthData string `json:"auth_data"`
	AgentID  uint   `json:"agent_id"`
	GroupID  uint   `json:"group_id"`
}

func NewServerHandler(db *store.DB) *ServerHandler {
	return &ServerHandler{db: db}
}

func (h *ServerHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.db.GetServers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

func (h *ServerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(uint(id))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	server := &model.Server{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		Username: req.Username,
		AuthType: req.AuthType,
		AuthData: req.AuthData,
		AgentID:  req.AgentID,
		GroupID:  req.GroupID,
	}

	if err := h.db.CreateServer(server); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(uint(id))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	server.Name = req.Name
	server.Host = req.Host
	server.Port = req.Port
	server.Username = req.Username
	server.AuthType = req.AuthType
	server.AuthData = req.AuthData
	server.AgentID = req.AgentID
	server.GroupID = req.GroupID

	if err := h.db.UpdateServer(server); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

func (h *ServerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteServer(uint(id)); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
