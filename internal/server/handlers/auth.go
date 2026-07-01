package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/neko233/uniops/internal/auth"
	"github.com/neko233/uniops/internal/store"
)

type AuthHandler struct {
	db         *store.DB
	jwtManager *auth.JWTManager
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func NewAuthHandler(db *store.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{db: db, jwtManager: jwtManager}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.Role = user.Role

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
