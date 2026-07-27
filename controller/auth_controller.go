package controller

import (
	"encoding/json"
	"errors"
	"fmt"

	"library-management-backend/dto"
	"library-management-backend/service"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in Register: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	var req dto.RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("[RegisterController] Invalid payload | error=%v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	err := ctrl.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEmail) {
			utils.ConflictResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.CreatedResponse(c, "User registered successfully", nil)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in Login: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	var req dto.LoginRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("[LoginController] Invalid payload | error=%v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			utils.UnauthorizedResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Login successful", res)
}
