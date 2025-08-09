package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/limistah/wallet-service/internal/auth"
)

// AuthMiddleware creates a middleware function for JWT authentication
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
				"error":   "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
				"error":   "authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Println(tokenString)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token is required",
				"error":   "empty token",
			})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// GetUserID extracts user ID from the Gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	if id, ok := userID.(uint); ok {
		return id, true
	}
	return 0, false
}

// GetUserEmail extracts user email from the Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}
	if emailStr, ok := email.(string); ok {
		return emailStr, true
	}
	return "", false
}
