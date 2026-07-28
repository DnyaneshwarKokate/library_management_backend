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

type BorrowController interface {
	BorrowBook(c *gin.Context)
	ReturnBook(c *gin.Context)
}

type borrowController struct {
	borrowService service.BorrowService
}

func NewBorrowController(borrowService service.BorrowService) BorrowController {
	return &borrowController{
		borrowService: borrowService,
	}
}

func (ctl *borrowController) BorrowBook(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("BorrowBook@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	UserId := c.GetHeader("auth_user_id")
	if UserId == "" {
		utils.UnauthorizedResponse(c, "Authorization failed: User ID is missing")
		return
	}
	userIDInt, err := strconv.Atoi(UserId)
	if err != nil {
		logrus.Error("Error converting user ID:", err)
		utils.ValidationResponse(c, "Invalid user ID")
		return
	}
	userID := uint(userIDInt)

	var req dto.BorrowBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("BorrowBook@Controller Invalid payload: %v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctl.borrowService.BorrowBook(req, userID)
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

func (ctl *borrowController) ReturnBook(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("ReturnBook@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	UserId := c.GetHeader("auth_user_id")
	if UserId == "" {
		utils.UnauthorizedResponse(c, "Authorization failed: User ID is missing")
		return
	}
	userIDInt, err := strconv.Atoi(UserId)
	if err != nil {
		logrus.Error("Error converting user ID:", err)
		utils.ValidationResponse(c, "Invalid user ID")
		return
	}
	userID := uint(userIDInt)

	recordUUID := c.Param("id")
	if recordUUID == "" {
		var req dto.ReturnBookRequest
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			_ = json.NewDecoder(c.Request.Body).Decode(&req)
			recordUUID = req.BorrowRecordUUID
		}
	}

	if recordUUID == "" {
		utils.BadRequestResponse(c, "invalid borrow record ID")
		return
	}

	res, err := ctl.borrowService.ReturnBook(recordUUID, userID)
	if err != nil {
		if errors.Is(err, service.ErrBorrowRecordNotFound) {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrAlreadyReturned) {
			utils.BadRequestResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrUnauthorizedReturn) {
			utils.ForbiddenResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Book returned successfully", res)
}
