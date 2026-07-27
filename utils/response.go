package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ResponseBody defines the common structure for JSON responses.
type ResponseBody struct {
	Message    string `json:"message"`
	StatusCode int    `json:"Status_code"`
	// DevMessage error       `json:"-"`
	DevMessage string      `json:"dev_message,omitempty"`
	Body       interface{} `json:"body,omitempty"`
}

type MiddlewareResponseBody struct {
	Message    string
	StatusCode int
	Body       struct{}
}

// Generic function to handle JSON responses
func sendJSONResponse(c *gin.Context, statusCode int, message string, body interface{}) {
	response := ResponseBody{
		Message:    message,
		StatusCode: statusCode,
		Body:       body,
	}
	c.JSON(statusCode, response)
}

// Generic function to abort with JSON response in middleware
func abortWithJSON(c *gin.Context, statusCode int, message string, body interface{}) {
	response := ResponseBody{
		StatusCode: statusCode,
		Message:    message,
		Body:       body,
	}
	c.AbortWithStatusJSON(statusCode, response)
}

// Success Response (200 OK)
func SuccessResponse(c *gin.Context, message string, body interface{}) {
	sendJSONResponse(c, http.StatusOK, message, body)
}

// Created Response (201 Created)
func CreatedResponse(c *gin.Context, message string, body interface{}) {
	sendJSONResponse(c, http.StatusCreated, message, body)
}

// Accepted Response (202 Accepted)
func AcceptedResponse(c *gin.Context, message string, body interface{}) {
	sendJSONResponse(c, http.StatusAccepted, message, body)
}

// No Content Response (204 No Content)
func NoContentResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusNoContent, message, nil)
}

// No Content Abort Response (204 No Content)
func NoContentAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusNoContent, message, nil)
}

// Validation Error Response (422 Unprocessable Entity)
func ValidationResponse(c *gin.Context, message string) {
	// sendJSONResponse(c, http.StatusUnprocessableEntity, message, map[string]interface{}{})
	sendJSONResponse(c, http.StatusUnprocessableEntity, message, map[string]interface{}{})
}

// Validation Abort Response (422 Unprocessable Entity)
func ValidationAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusUnprocessableEntity, message, map[string]interface{}{})
}

// Bad Request Response (400 Bad Request)
func BadRequestResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusBadRequest, message, nil)
}

// Bad Request Abort Response (400 Bad Request)
func BadRequestAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusBadRequest, message, nil)
}

// Unauthorized Response (401 Unauthorized)
func UnauthorizedResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusUnauthorized, message, nil)
}

// Unauthorized Abort Response (401 Unauthorized)
func UnauthorizedAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusUnauthorized, message, nil)
}

// Forbidden Response (403 Forbidden)
func ForbiddenResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusForbidden, message, nil)
}

// Forbidden Abort Response (403 Forbidden)
func ForbiddenAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusForbidden, message, nil)
}

// Not Found Response (404 Not Found)
func NotFoundResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusNotFound, message, map[string]interface{}{})
}

// Not Found Abort Response (404 Not Found)
func NotFoundAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusNotFound, message, map[string]interface{}{})
}

// Internal Server Error Response (500 Internal Server Error)
func InternalServerErrorResponse(c *gin.Context, err error) {
	response := ResponseBody{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal Server Error",
		DevMessage: err.Error(),
	}
	c.JSON(http.StatusInternalServerError, response)
}

// Internal Server Error Abort Response (500 Internal Server Error)
func InternalServerErrorAbortWithJSON(c *gin.Context, err error) {
	response := ResponseBody{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal Server Error",
		DevMessage: err.Error(),
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, response)
}

// Service Unavailable Response (503 Service Unavailable)
func ServiceUnavailableResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusServiceUnavailable, message, nil)
}

// Service Unavailable Abort Response (503 Service Unavailable)
func ServiceUnavailableAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusServiceUnavailable, message, nil)
}

// Gateway Timeout Response (504 Gateway Timeout)
func GatewayTimeoutResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusGatewayTimeout, message, nil)
}

// Gateway Timeout Abort Response (504 Gateway Timeout)
func GatewayTimeoutAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusGatewayTimeout, message, nil)
}

// Internal Server Error with custom message (500 Internal Server Error)
func InternalServerErrorWithMessage(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusInternalServerError, message, map[string]interface{}{})
}

// Internal Server Error Abort with custom message (500 Internal Server Error)
func InternalServerErrorAbortWithMessage(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusInternalServerError, message, map[string]interface{}{})
}

// Too Many Requests Response (429 Too Many Requests)
func TooManyRequestsResponse(c *gin.Context, message string) {
	sendJSONResponse(c, http.StatusTooManyRequests, message, nil)
}

// Too Many Requests Abort Response (429 Too Many Requests)
func TooManyRequestsAbortWithJSON(c *gin.Context, message string) {
	abortWithJSON(c, http.StatusTooManyRequests, message, nil)
}

func RollbackWithError(tx *gorm.DB, errMsg string, err error) error {
	if tx != nil {
		tx.Rollback()
	}
	logrus.Errorf("%s: %v", errMsg, err)
	return fmt.Errorf("%s: %v", errMsg, err)
}

func RespondWithError(c *gin.Context, code int, message string) {
	response := MiddlewareResponseBody{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
	c.AbortWithStatusJSON(code, response)
}

func BadRequestWithObject(c *gin.Context, body interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  false,
		"message": body,
	})
}
func ConflictResponse(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"message":     message,
		"status_code": http.StatusConflict,
		"body":        gin.H{},
	})
}
