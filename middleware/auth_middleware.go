package middleware

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"library-management-backend/constants"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const (
	ContextUserID   = "user_id"
	ContextUserUUID = "user_uuid"
	ContextUserRole = "user_role"
)

func TokenAuthentication(jwtSecret ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			utils.UnauthorizedAbortWithJSON(c, "Missing token.")
			return
		}

		secret := ""
		if len(jwtSecret) > 0 && jwtSecret[0] != "" {
			secret = jwtSecret[0]
		} else {
			secret = os.Getenv("JWT_SECRET")
			if secret == "" {
				secret = "super_secret_jwt_key_library_app"
			}
		}

		tokenStringClean := strings.TrimSpace(tokenString)
		parts := strings.SplitN(tokenStringClean, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenStringClean = strings.TrimSpace(parts[1])
		}

		claims, err := utils.ValidateToken(tokenStringClean, secret)
		if err != nil {
			utils.UnauthorizedAbortWithJSON(c, "Unauthorized Access.")
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserUUID, claims.UserUUID)
		c.Set(ContextUserRole, claims.Role)

		c.Request.Header.Set("user_id", fmt.Sprintf("%v", claims.UserUUID))
		c.Request.Header.Set("auth_user_id", fmt.Sprintf("%v", claims.UserID))
		c.Request.Header.Set("user_type", fmt.Sprintf("%v", claims.Role))
		c.Next()
	}
}

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return TokenAuthentication(jwtSecret)
}

func TokenAuthenticationHandler(c *gin.Context) {
	fmt.Println("app_env : ", os.Getenv("APP_ENV"))
	authorization := c.Request.Header.Get("Authorization")
	if authorization != "" {
		tokenValidClaims, tokenErr := TokenValid(authorization)
		fmt.Println(tokenValidClaims)
		logrus.Info("Tokendata: ", tokenValidClaims)
		if tokenErr != nil {
			utils.UnauthorizedAbortWithJSON(c, "Invalid Token.")
			return
		}
		logrus.Info("tokenValidClaims", tokenValidClaims)
		if ssoId, ok := tokenValidClaims["sub"].(string); ok {
			c.Request.Header.Set("sub_id", ssoId)
		}
		c.Next()
	} else {
		utils.UnauthorizedAbortWithJSON(c, "Unauthorized Access.")
		return
	}
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	pubKeyFileExt := "pem"
	if os.Getenv("APP_ENV") == "production" {
		pubKeyFileExt = "pub"
	}
	keyPath := "storage/ssl/" + os.Getenv("APP_ENV") + "/jwtRSA256-public-" + os.Getenv("APP_ENV") + "." + pubKeyFileExt
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		logrus.Info("testkeyData err", err)
		return nil, err
	}
	logrus.Info("VerifyToken@Path : ", keyPath)

	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		logrus.Info("key err", err)
		return nil, err
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	return parsedToken, nil
}

func TokenValid(tokenString string) (jwt.MapClaims, error) {
	token, err := VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token claims")
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
