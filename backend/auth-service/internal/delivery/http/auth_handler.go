package http

import (
	"encoding/json"
	"net/http"
	"time"

	"go-forum-project/auth-service/internal/usecase"
)

type AuthHandler struct {
	AuthUC *usecase.AuthUseCase
}
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessExp    time.Time `json:"access_exp"`
	RefreshExp   time.Time `json:"refresh_exp"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := h.AuthUC.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		AccessExp:    tokens.AccessExpiresAt,
		RefreshExp:   tokens.RefreshExpiresAt,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		http.Error(w, "refresh token required", http.StatusUnauthorized)
	}

	newTokens, err := h.AuthUC.RefreshTokens(refreshToken)
	if err != nil {
		http.Error(w, "refresh failed", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		AccessExp:    newTokens.AccessExpiresAt,
		RefreshExp:   newTokens.RefreshExpiresAt,
	})
}
