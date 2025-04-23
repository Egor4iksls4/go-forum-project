package http

import (
	"encoding/json"
	"go-forum-project/auth-service/internal/usecase"
	"net/http"
)

type AuthHandler struct {
	AuthUC *usecase.AuthUseCase
}
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool   `json:"success"`
	Role    string `json:"role"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.AuthUC.Login(req.Username, req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(loginResponse{Success: false})
		return
	}

	json.NewEncoder(w).Encode(loginResponse{
		Success: true,
		Role:    user.Role,
	})
}
