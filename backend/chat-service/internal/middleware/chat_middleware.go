package middleware

import (
	"context"
	"github.com/gorilla/websocket"
	"go-forum-project/chat-service/internal/client"
	"log"
	"net/http"
	"strings"
)

type WebSocketAuth struct {
	authClient *client.AuthClient
	upgrader   *websocket.Upgrader
}

func NewWebSocketAuth(authClient *client.AuthClient) *WebSocketAuth {
	return &WebSocketAuth{
		authClient: authClient,
		upgrader: &websocket.Upgrader{
			Subprotocols:    []string{"Bearer"},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				allowedOrigins := []string{
					"http://localhost:3000",
				}
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				return false
			},
		},
	}
}

func (ws *WebSocketAuth) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(wt http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Upgrade"), "websocket") {
			accessToken := extractTokenFromRequest(r)
			if accessToken == "" {
				http.Error(wt, "Authorization token error", http.StatusUnauthorized)
				return
			}

			username, valid, err := ws.authClient.ValidateToken(r.Context(), accessToken)
			if err != nil || !valid {
				http.Error(wt, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "username", username)
			next(wt, r.WithContext(ctx))
			return
		}

		accessToken := extractTokenFromRequest(r)
		if accessToken == "" {
			http.Error(wt, "Authorization token error", http.StatusUnauthorized)
			return
		}

		username, valid, err := ws.authClient.ValidateToken(r.Context(), accessToken)
		if err != nil {
			http.Error(wt, "Token validation error", http.StatusUnauthorized)
			return
		}

		if !valid {
			refreshToken := r.Header.Get("X-Refresh-Token")
			if refreshToken == "" {
				http.Error(wt, "Refresh token required", http.StatusUnauthorized)
				return
			}

			newTokens, err := ws.authClient.Refresh(r.Context(), refreshToken)
			if err != nil {
				log.Printf("Refresh failed: %v", err)
				http.Error(wt, "Refresh failed", http.StatusUnauthorized)
				return
			}

			wt.Header().Set("New-Access-Token", newTokens.AccessToken)
			wt.Header().Set("New-Refresh-Token", newTokens.RefreshToken)

			username, valid, err = ws.authClient.ValidateToken(r.Context(), newTokens.AccessToken)
			if err != nil || !valid {
				http.Error(wt, "Token validation error after refresh", http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), "username", username)

		if strings.Contains(r.Header.Get("Upgrade"), "websocket") {
			conn, err := ws.upgrader.Upgrade(wt, r, nil)
			if err != nil {
				log.Println("WebSocket upgrade error:", err)
				return
			}

			conn.WriteJSON(map[string]interface{}{
				"type":     "auth",
				"username": username,
			})

			next(wt, r.WithContext(ctx))
			return
		}

		next(wt, r.WithContext(ctx))
	}
}

func extractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			return tokenParts[1]
		}
	}

	return r.URL.Query().Get("token")
}
