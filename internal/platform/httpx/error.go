// Package httpx holded the HTTP plumbing shared by every feature module.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the error shape returned bu every endpoint.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type errorResponse struct {
	Error ErrorBody `json:"error"`
}

// AbortWithError stops the handler chain and write the standard error body.
func AbortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorResponse{
		Error: ErrorBody{Code: code, Message: message, RequestID: GetRequestID(c)},
	})
}

// NotFound is the handler from unknown routes.
func NotFound(c *gin.Context) {
	AbortWithError(c, http.StatusNotFound, "NOT_FOUND", "The requested resource not found")
}

// MethodNotAllowed is the handler for known routes called with the wrong methods.
func MethodNotAllowed(c *gin.Context) {
	AbortWithError(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "This method is not allowed for the requested resource")
}
