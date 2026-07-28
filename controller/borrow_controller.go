package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"library-management-backend/dto"
	"library-management-backend/service"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type BorrowController struct {
	borrowService service.BorrowService
}

func NewBorrowController(borrowService service.BorrowService) *BorrowController {
	return &BorrowController{
		borrowService: borrowService,
	}
}

func (ctrl *BorrowController) BorrowBook(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in BorrowBook: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	userIDStr := c.GetHeader("auth_user_id")
	id, _ := strconv.ParseUint(userIDStr, 10, 32)
	userID := uint(id)
	if userID == 0 {
		utils.UnauthorizedResponse(c, "User context missing")
		return
	}

	var req dto.BorrowBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("[BorrowBookController] Invalid payload | error=%v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.borrowService.BorrowBook(req, userID)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) || errors.Is(err, service.ErrUserNotFound) {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrBookOutOfStock) ||
			errors.Is(err, service.ErrBookAlreadyBorrowed) ||
			errors.Is(err, service.ErrBorrowLimitExceeded) {
			utils.BadRequestResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.CreatedResponse(c, "Book borrowed successfully", res)
}
