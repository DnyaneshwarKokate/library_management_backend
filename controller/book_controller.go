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
	GetBooks(c *gin.Context)
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

func (ctl *bookController) GetBooks(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("GetBooks@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	var filter dto.BookFilter
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		_ = json.NewDecoder(c.Request.Body).Decode(&filter)
	}
	if filter.Limit <= 0 && filter.Offset <= 0 && filter.Search == "" && filter.Category == "" {
		_ = c.ShouldBindQuery(&filter)
	}

	res, err := ctl.bookService.GetBooks(filter)
	if err != nil {
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Books fetched successfully", res)
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
