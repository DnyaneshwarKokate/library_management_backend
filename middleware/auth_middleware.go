package middleware

import (
	"net/http"
	"strings"

	"library-management-backend/constants"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID   = "user_id"
	ContextUserUUID = "user_uuid"
	ContextUserRole = "user_role"
)

// AuthMiddleware validates incoming JWT tokens and injects claims into Gin context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.SendError(c, http.StatusUnauthorized, "Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			utils.SendError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "Invalid or expired authorization token", err.Error())
			c.Abort()
			return
		}

		// Inject user context
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserUUID, claims.UserUUID)
		c.Set(ContextUserRole, claims.Role)

		c.Next()
	}
}

// RequireRole restricts access to endpoints based on allowed user roles.
func RequireRole(allowedRoles ...constants.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextUserRole)
		if !exists {
			utils.SendError(c, http.StatusUnauthorized, "User context missing", nil)
			c.Abort()
			return
		}

		userRole, ok := roleVal.(constants.Role)
		if !ok {
			utils.SendError(c, http.StatusInternalServerError, "Invalid role context type", nil)
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.SendError(c, http.StatusForbidden, "Access denied: insufficient permissions", nil)
		c.Abort()
	}
}
