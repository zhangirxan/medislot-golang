package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"medislot/internal/models"
	"medislot/pkg/utils"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "role"
	ContextKeyEmail  = "email"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing or malformed Authorization header",
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			return
		}

		userID := claims.Subject
		if userID == "" {
			userID = claims.UserID
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, string(claims.Role))
		c.Next()
	}
}

func RequireRole(allowed ...models.Role) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		set[string(r)] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get(ContextKeyRole)
		if _, ok := set[role.(string)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   models.ErrForbidden.Error(),
			})
			return
		}
		c.Next()
	}
}

func GetCallerID(c *gin.Context) string {
	id, exists := c.Get(ContextKeyUserID)
	if !exists {
		return ""
	}
	str, ok := id.(string)
	if !ok {
		return ""
	}
	return str
}

func GetCallerRole(c *gin.Context) models.Role {
	r, exists := c.Get(ContextKeyRole)
	if !exists {
		return ""
	}
	str, ok := r.(string)
	if !ok {
		return ""
	}
	return models.Role(str)
}
