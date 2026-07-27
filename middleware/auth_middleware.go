package middleware

import (
	"errors"
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

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedAbortWithJSON(c, "Authorization header is required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && strings.EqualFold(parts[0], "Bearer")) {
			utils.UnauthorizedAbortWithJSON(c, "Authorization header format must be Bearer {token}")
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			utils.UnauthorizedAbortWithJSON(c, "Invalid or expired authorization token")
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserUUID, claims.UserUUID)
		c.Set(ContextUserRole, claims.Role)

		c.Next()
	}
}

func RequireRole(allowedRoles ...constants.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextUserRole)
		if !exists {
			utils.UnauthorizedAbortWithJSON(c, "User context missing")
			return
		}

		userRole, ok := roleVal.(constants.Role)
		if !ok {
			utils.InternalServerErrorAbortWithJSON(c, errors.New("invalid role context type"))
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.ForbiddenAbortWithJSON(c, "Access denied: insufficient permissions")
	}
}
