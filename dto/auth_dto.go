package dto

import (
	"time"

	"library-management-backend/constants"
)

type RegisterRequest struct {
	Name     string         `json:"name" binding:"required,min=2,max=100"`
	Email    string         `json:"email" binding:"required,email,max=100"`
	Password string         `json:"password" binding:"required,min=6,max=100"`
	Role     constants.Role `json:"role" binding:"omitempty,oneof=ADMIN MEMBER"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID        uint           `json:"id"`
	UUID      string         `json:"uuid"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Role      constants.Role `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
