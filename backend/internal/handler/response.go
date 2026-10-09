package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard error shape per constitution §6.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains the error code and message.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSONError writes a standard error response.
func JSONError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// AbortWithError writes a standard error response and aborts the middleware chain.
func AbortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// NotFoundHandler returns a JSON 404 for unmatched API routes.
func NotFoundHandler(c *gin.Context) {
	JSONError(c, http.StatusNotFound, "NOT_FOUND", "The requested resource was not found.")
}
