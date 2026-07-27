package controller

import (
	"errors"
	"net/http"

	"library-management-backend/dto"
	"library-management-backend/service"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	err := ctrl.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEmail) {
			utils.SendError(c, http.StatusConflict, err.Error(), nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to register user", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "User registered successfully", nil)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	res, err := ctrl.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			utils.SendError(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		utils.SendError(c, http.StatusInternalServerError, "Failed to log in", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Login successful", res)
}
