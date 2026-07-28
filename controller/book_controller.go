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

type BookController interface {
	CreateBook(c *gin.Context)
	GetBooksList(c *gin.Context)
	GetBookByUUID(c *gin.Context)
	UpdateBook(c *gin.Context)
	DeleteBook(c *gin.Context)
}

type bookController struct {
	bookService service.BookService
}

func NewBookController(bookService service.BookService) BookController {
	return &bookController{
		bookService: bookService,
	}
}

func (ctl *bookController) CreateBook(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("CreateBook@Controller panic: %v", panicInfo)
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
	adminID := uint(userIDInt)

	var req dto.CreateBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("CreateBook@Controller Invalid payload: %v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctl.bookService.CreateBook(req, adminID)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateISBN) {
			utils.ConflictResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.CreatedResponse(c, "Book created successfully", res)
}

func (ctl *bookController) GetBooksList(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("GetBooksList@panic:", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	var req dto.BookFilter
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			logrus.Error("Invalid request body:", err)
			utils.ValidationResponse(c, "Invalid request data")
			return
		}
	} else {
		_ = c.ShouldBindQuery(&req)
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	response, total, filtered, err := ctl.bookService.GetBooksList(req)
	if err != nil {
		logrus.Error("GetBooksList@Error:", err)
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Books list fetched successfully", map[string]interface{}{
		"data":         response,
		"filter_count": filtered,
		"total_count":  total,
	})
}

func (ctl *bookController) GetBookByUUID(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("GetBookByUUID@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	uuid := c.Param("uuid")
	if uuid == "" {
		var payload struct {
			UUID string `json:"uuid"`
		}
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			_ = json.NewDecoder(c.Request.Body).Decode(&payload)
			uuid = payload.UUID
		}
	}

	if uuid == "" {
		utils.BadRequestResponse(c, "invalid book UUID")
		return
	}

	res, err := ctl.bookService.GetBookByUUID(uuid)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Book details fetched successfully", res)
}

func (ctl *bookController) UpdateBook(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("UpdateBook@Controller panic: %v", panicInfo)
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
	adminID := uint(userIDInt)

	uuid := c.Param("uuid")
	var req dto.UpdateBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("UpdateBook@Controller Invalid payload: %v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if uuid == "" {
		uuid = req.UUID
	}

	if uuid == "" {
		utils.BadRequestResponse(c, "invalid book UUID")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctl.bookService.UpdateBook(uuid, req, adminID)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidAvailableCopies) {
			utils.BadRequestResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Book updated successfully", res)
}

func (ctl *bookController) DeleteBook(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("DeleteBook@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	UserId := c.GetHeader("auth_user_id")
	if UserId == "" {
		utils.UnauthorizedResponse(c, "Authorization failed: User ID is missing")
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		var payload struct {
			UUID string `json:"uuid"`
		}
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			_ = json.NewDecoder(c.Request.Body).Decode(&payload)
			uuid = payload.UUID
		}
	}

	if uuid == "" {
		utils.BadRequestResponse(c, "invalid book UUID")
		return
	}

	err := ctl.bookService.DeleteBook(uuid)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			utils.NotFoundResponse(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrActiveBorrowRecordsExist) {
			utils.BadRequestResponse(c, err.Error())
			return
		}
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Book deleted successfully", nil)
}
