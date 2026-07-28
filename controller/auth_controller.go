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

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type authController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) AuthController {
	return &authController{
		authService: authService,
	}
}

func (ctl *authController) Register(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("Register@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	var req dto.RegisterRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("Register@Controller Invalid payload: %v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	err := ctl.authService.Register(req)
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

func (ctl *authController) Login(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("Login@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	var req dto.LoginRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("Login@Controller Invalid payload: %v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctl.authService.Login(req)
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
