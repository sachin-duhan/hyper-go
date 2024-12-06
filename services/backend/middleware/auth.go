package middleware

import (
	"net/http"
	"strings"

	"go-turbo/pkg/auth"
	"go-turbo/pkg/database"
	"go-turbo/pkg/models"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	db *database.Database
}

func NewAuthMiddleware(db *database.Database) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization header"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user ID to context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		user, err := models.GetUserByID(c.Request.Context(), m.db.Pool, userID.(uint))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
