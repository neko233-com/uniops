package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/neko233/uniops/internal/sftp"
	"github.com/neko233/uniops/internal/store"
)

type FileManagerHandler struct {
	db      *store.DB
	handler *sftp.Handler
}

type ListFilesRequest struct {
	Path string `json:"path"`
}

type UploadRequest struct {
	Path string `json:"path"`
}

func NewFileManagerHandler(db *store.DB) *FileManagerHandler {
	return &FileManagerHandler{
		db:      db,
		handler: sftp.NewHandler(),
	}
}

func (h *FileManagerHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	var req ListFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Path = "/"
	}

	files, err := h.handler.ListFiles(server.Host, server.Port, server.Username, server.AuthData, req.Path)
	if err != nil {
		http.Error(w, "failed to list files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *FileManagerHandler) Download(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := chi.URLParam(r, "*")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	file, err := h.handler.DownloadFile(server.Host, server.Port, server.Username, server.AuthData, "/"+remotePath)
	if err != nil {
		http.Error(w, "download failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment")
	io.Copy(w, file)
}

func (h *FileManagerHandler) Upload(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := r.URL.Query().Get("path")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	if err := h.handler.UploadFile(server.Host, server.Port, server.Username, server.AuthData, remotePath, r.Body); err != nil {
		http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *FileManagerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	remotePath := chi.URLParam(r, "*")
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	if err := h.handler.DeleteFile(server.Host, server.Port, server.Username, server.AuthData, "/"+remotePath); err != nil {
		http.Error(w, "delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *FileManagerHandler) Mkdir(w http.ResponseWriter, r *http.Request) {
	serverID, _ := strconv.ParseUint(chi.URLParam(r, "serverId"), 10, 32)
	
	server, err := h.db.GetServer(uint(serverID))
	if err != nil {
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.handler.Mkdir(server.Host, server.Port, server.Username, server.AuthData, req.Path); err != nil {
		http.Error(w, "mkdir failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}