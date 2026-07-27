package service

import (
	"errors"
	"os"

	"library-management-backend/constants"
	"library-management-backend/dto"
	"library-management-backend/model"
	"library-management-backend/repository"
	"library-management-backend/utils"
)

var (
	ErrDuplicateEmail    = errors.New("email is already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
)

type AuthService interface {
	Register(req dto.RegisterRequest) error
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "super_secret_jwt_key_library_app"
	}
	return secret
}

func (s *authService) Register(req dto.RegisterRequest) error {
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}
	role := req.Role
	if role == "" {
		role = constants.RoleMember
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	user := &model.User{
		UUID:         utils.WithoutHypenGenUUID(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	_, err = s.userRepo.StoreUser(user)
	return err
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredential
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredential
	}
	expiryHours := 24
	token, err := utils.GenerateToken(user.ID, user.UUID, user.Role, s.getJWTSecret(), expiryHours)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			UUID:      user.UUID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
