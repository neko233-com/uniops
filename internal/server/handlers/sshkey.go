package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/ssh"
	"github.com/neko233/uniops/internal/store"
)

type SSHKeyHandler struct {
	db         *store.DB
	keyManager *ssh.KeyManager
	deployer   *ssh.Deployer
}

type GenerateKeyRequest struct {
	ServerID uint   `json:"server_id"`
	Bits     int    `json:"bits"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type DeployKeyRequest struct {
	KeyID uint `json:"key_id"`
}

func NewSSHKeyHandler(db *store.DB) *SSHKeyHandler {
	return &SSHKeyHandler{
		db:         db,
		keyManager: ssh.NewKeyManager(),
		deployer:   ssh.NewDeployer(),
	}
}

func (h *SSHKeyHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	server, err := h.db.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	bits := req.Bits
	if bits == 0 {
		bits = 2048
	}

	privateKey, publicKey, fingerprint, err := h.keyManager.GenerateKeyPair(bits)
	if err != nil {
		http.Error(w, "failed to generate key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sshKey := &model.SSHKey{
		Name:        server.Name + "-key",
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Fingerprint: fingerprint,
		ServerID:    req.ServerID,
		Status:      "pending",
	}

	if err := h.db.CreateSSHKey(sshKey); err != nil {
		http.Error(w, "failed to save key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sshKey)
}

func (h *SSHKeyHandler) Deploy(w http.ResponseWriter, r *http.Request) {
	var req DeployKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	key, err := h.db.GetSSHKey(req.KeyID)
	if err != nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	server, err := h.db.GetServer(key.ServerID)
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	if err := h.deployer.DeployKey(server.Host, server.Port, server.Username, server.AuthData, key.PublicKey); err != nil {
		http.Error(w, "deploy failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	key.Status = "deployed"
	h.db.UpdateSSHKey(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "deployed",
		"key":    key,
	})
}

func (h *SSHKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)

	keys, err := h.db.GetSSHKeysByServer(uint(serverID))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *SSHKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	key, err := h.db.GetSSHKey(uint(id))
	if err != nil {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}
