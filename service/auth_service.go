package service

import (
	"errors"
	"os"

	"library-management-backend/constants"
	"library-management-backend/dto"
	"library-management-backend/model"
	"library-management-backend/repository"
	"library-management-backend/utils"

	"github.com/sirupsen/logrus"
)

var (
	ErrDuplicateEmail    = errors.New("email is already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrUserInactive      = errors.New("user account is inactive")
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
	logrus.Info("Register@Service Started")

	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		logrus.Errorf("Register@Service ExistsByEmail Error: %v", err)
		return err
	}
	if exists {
		logrus.Warnf("Register@Service Duplicate Email: %s", req.Email)
		return ErrDuplicateEmail
	}

	role := req.Role
	if role == "" {
		role = constants.RoleMember
	}
	status := req.Status
	if status == "" {
		status = constants.StatusActive
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logrus.Errorf("Register@Service HashPassword Error: %v", err)
		return err
	}

	user := &model.User{
		UUID:         utils.WithoutHypenGenUUID(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         role,
		Status:       status,
	}

	storedUser, err := s.userRepo.StoreUser(nil, user)
	if err != nil {
		logrus.Errorf("Register@Service StoreUser Error: %v", err)
		return err
	}

	logrus.Infof("Register@Service Completed successfully for User ID: %d", storedUser.ID)
	return nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	logrus.Info("Login@Service Started")

	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		logrus.Errorf("Login@Service FindByEmail Error: %v", err)
		return nil, err
	}
	if user == nil {
		logrus.Warnf("Login@Service Invalid Email: %s", req.Email)
		return nil, ErrInvalidCredential
	}
	if user.Status == constants.StatusInactive {
		logrus.Warnf("Login@Service Inactive User Account ID: %d", user.ID)
		return nil, ErrUserInactive
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		logrus.Warnf("Login@Service Password Check Failed for Email: %s", req.Email)
		return nil, ErrInvalidCredential
	}

	expiryHours := 24
	token, err := utils.GenerateToken(user.ID, user.UUID, user.Role, s.getJWTSecret(), expiryHours)
	if err != nil {
		logrus.Errorf("Login@Service GenerateToken Error: %v", err)
		return nil, err
	}

	logrus.Infof("Login@Service Completed successfully for User ID: %d", user.ID)
	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			UUID:      user.UUID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
