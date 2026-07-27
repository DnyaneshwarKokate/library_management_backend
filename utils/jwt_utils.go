package utils

import (
	"errors"
	"fmt"
	"time"

	"library-management-backend/constants"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   uint           `json:"user_id"`
	UserUUID string         `json:"user_uuid"`
	Role     constants.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, userUUID string, role constants.Role, secret string, expiryHours int) (string, error) {
	if expiryHours <= 0 {
		expiryHours = 24
	}

	claims := &JWTClaims{
		UserID:   userID,
		UserUUID: userUUID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "library-management-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
