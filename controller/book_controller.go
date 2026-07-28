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

type BookController struct {
	bookService service.BookService
}

func NewBookController(bookService service.BookService) *BookController {
	return &BookController{
		bookService: bookService,
	}
}

func (ctrl *BookController) CreateBook(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in CreateBook: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	var req dto.CreateBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("[CreateBookController] Invalid payload | error=%v", err)
		utils.ValidationResponse(c, "invalid request data")
		return
	}

	if validationResp := utils.ValidateRequest(c, req); validationResp != nil {
		utils.ValidationResponse(c, validationResp.(string))
		return
	}

	userIDStr := c.GetHeader("auth_user_id")
	id, _ := strconv.ParseUint(userIDStr, 10, 32)
	adminID := uint(id)
	res, err := ctrl.bookService.CreateBook(req, adminID)
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

func (ctrl *BookController) GetBooks(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in GetBooks: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	var filter dto.BookFilter
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		_ = json.NewDecoder(c.Request.Body).Decode(&filter)
	}
	if filter.Limit <= 0 && filter.Offset <= 0 && filter.Search == "" && filter.Category == "" {
		_ = c.ShouldBindQuery(&filter)
	}

	res, err := ctrl.bookService.GetBooks(filter)
	if err != nil {
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Books fetched successfully", res)
}

func (ctrl *BookController) GetBookByUUID(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in GetBookByUUID: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
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

	res, err := ctrl.bookService.GetBookByUUID(uuid)
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

func (ctrl *BookController) UpdateBook(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in UpdateBook: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
		}
	}()

	uuid := c.Param("uuid")
	var req dto.UpdateBookRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		logrus.Errorf("[UpdateBookController] Invalid payload | error=%v", err)
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

	userIDStr := c.GetHeader("auth_user_id")
	id, _ := strconv.ParseUint(userIDStr, 10, 32)
	adminID := uint(id)
	res, err := ctrl.bookService.UpdateBook(uuid, req, adminID)
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

func (ctrl *BookController) DeleteBook(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("Recovered in DeleteBook: ", r)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", r))
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

	err := ctrl.bookService.DeleteBook(uuid)
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
