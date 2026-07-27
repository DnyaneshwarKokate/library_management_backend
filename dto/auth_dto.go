package dto

import (
	"time"

	"library-management-backend/constants"
)

type RegisterRequest struct {
	Name     string               `json:"name" validate:"required,min=2,max=100" binding:"required,min=2,max=100"`
	Email    string               `json:"email" validate:"required,email,max=100" binding:"required,email,max=100"`
	Password string               `json:"password" validate:"required,min=6,max=100" binding:"required,min=6,max=100"`
	Role     constants.Role       `json:"role" validate:"omitempty,oneof=ADMIN MEMBER" binding:"omitempty,oneof=ADMIN MEMBER"`
	Status   constants.UserStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" binding:"required,email"`
	Password string `json:"password" validate:"required" binding:"required"`
}

type UserResponse struct {
	ID        uint                 `json:"id"`
	UUID      string               `json:"uuid"`
	Name      string               `json:"name"`
	Email     string               `json:"email"`
	Role      constants.Role       `json:"role"`
	Status    constants.UserStatus `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
