package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination Pagination  `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}
func CreatedResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, status int, message string, err error) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}

	c.JSON(status, Response{
		Success: false,
		Message: message,
		Error:   errorMessage,
	})
}

func BadRequestResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusBadRequest, message, err)
}

func NotFoundResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusNotFound, message, err)
}

func InternalServerErrorResponse(c *gin.Context, message string, err error) {
	ErrorResponse(c, http.StatusInternalServerError, message, err)
}

func UnathorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, nil)
}

func PaginatedSuccessResponse(c *gin.Context, message string, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}
