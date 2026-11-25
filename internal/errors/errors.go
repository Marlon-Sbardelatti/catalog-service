package apierrors

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	StatusCode int               `json:"-"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

func newError(statusCode int, code, message string, details map[string]string) APIError {
	return APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

func BadRequest(message string, details map[string]string) APIError {
	return newError(http.StatusBadRequest, "REQUEST_INVALID", message, details)
}

func UnprocessableEntity(code, message string) APIError {
	return newError(http.StatusUnprocessableEntity, code, message, nil)
}

func NotFound(code, message string) APIError {
	return newError(http.StatusNotFound, code, message, nil)
}

func Conflict(code, message string) APIError {
	return newError(http.StatusConflict, code, message, nil)
}

func MethodNotAllowed() APIError {
	return newError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
}

func Internal() APIError {
	return newError(
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"an unexpected error occurred",
		nil,
	)
}

// Write is the only place that responds with an error and the only place that
// logs one: server errors log their cause here, client errors are left to the caller.
func Write(c *gin.Context, apiErr APIError, cause error) {
	if apiErr.StatusCode >= http.StatusInternalServerError {
		slog.Error(
			"request failed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", apiErr.StatusCode,
			"error", cause,
		)
	}

	c.JSON(apiErr.StatusCode, apiErr)
}
