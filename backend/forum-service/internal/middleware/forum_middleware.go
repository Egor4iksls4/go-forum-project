package middleware

import (
	"github.com/gin-gonic/gin"
	"go-forum-project/forum-service/internal/client"
	"net/http"
	"strings"
)

func AuthMiddleware(authClient *client.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := extractTokenFromHeader(c)
		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		refreshToken := c.GetHeader("X-Refresh-Token")
		if refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
			return
		}

		username, valid, err := authClient.ValidateToken(c.Request.Context(), accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation error"})
			return
		}

		if valid {
			c.Set("username", username)
			c.Next()
			return
		}

		_, err = authClient.Refresh(c.Request.Context(), refreshToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":         "Session expired",
				"should_logout": true,
			})
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

func extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return ""
	}

	return tokenParts[1]
}
