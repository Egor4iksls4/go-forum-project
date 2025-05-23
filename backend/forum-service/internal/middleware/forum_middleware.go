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

		username, valid, err := authClient.ValidateToken(c.Request.Context(), accessToken)
		if err != nil {
			refreshToken := c.GetHeader("X-Refresh-Token")
			if refreshToken == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation error and no refresh token"})
				return
			}

			// Пытаемся обновить токен
			newTokens, refreshErr := authClient.Refresh(c.Request.Context(), refreshToken)
			if refreshErr != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh failed"})
				return
			}

			c.Header("New-Access-Token", newTokens.AccessToken)
			c.Header("New-Refresh-Token", newTokens.RefreshToken)

			newUsername, _, validationErr := authClient.ValidateToken(c.Request.Context(), newTokens.AccessToken)
			if validationErr != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "New token validation failed"})
				return
			}

			c.Set("username", newUsername)
			c.Next()
			return
		}

		if valid {
			c.Set("username", username)
			c.Next()
			return
		}

		refreshToken := c.GetHeader("X-Refresh-Token")
		if refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
			return
		}

		newTokens, err := authClient.Refresh(c.Request.Context(), refreshToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh failed"})
			return
		}

		c.Header("New-Access-Token", newTokens.AccessToken)
		c.Header("New-Refresh-Token", newTokens.RefreshToken)

		newUsername, _, err := authClient.ValidateToken(c.Request.Context(), newTokens.AccessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed"})
			return
		}

		c.Set("username", newUsername)
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
