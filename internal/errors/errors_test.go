package apierrors

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		err        APIError
		statusCode int
		code       string
		message    string
	}{
		{
			name:       "bad request",
			err:        BadRequest("invalid product id", nil),
			statusCode: http.StatusBadRequest,
			code:       "REQUEST_INVALID",
			message:    "invalid product id",
		},
		{
			name:       "unprocessable entity",
			err:        UnprocessableEntity("PRODUCT_INVALID", "invalid product data"),
			statusCode: http.StatusUnprocessableEntity,
			code:       "PRODUCT_INVALID",
			message:    "invalid product data",
		},
		{
			name:       "not found",
			err:        NotFound("PRODUCT_NOT_FOUND", "product not found"),
			statusCode: http.StatusNotFound,
			code:       "PRODUCT_NOT_FOUND",
			message:    "product not found",
		},
		{
			name:       "conflict",
			err:        Conflict("PRODUCT_ALREADY_EXISTS", "product already exists"),
			statusCode: http.StatusConflict,
			code:       "PRODUCT_ALREADY_EXISTS",
			message:    "product already exists",
		},
		{
			name:       "method not allowed",
			err:        MethodNotAllowed(),
			statusCode: http.StatusMethodNotAllowed,
			code:       "METHOD_NOT_ALLOWED",
			message:    "method not allowed",
		},
		{
			name:       "internal",
			err:        Internal(),
			statusCode: http.StatusInternalServerError,
			code:       "INTERNAL_SERVER_ERROR",
			message:    "an unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", tt.err.StatusCode, tt.statusCode)
			}
			if tt.err.Code != tt.code {
				t.Errorf("Code = %q, want %q", tt.err.Code, tt.code)
			}
			if tt.err.Message != tt.message {
				t.Errorf("Message = %q, want %q", tt.err.Message, tt.message)
			}
		})
	}
}

func TestWriteLogsOnlyServerErrors(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		apiErr  APIError
		wantLog bool
	}{
		{name: "client error is not logged", apiErr: BadRequest("invalid product id", nil), wantLog: false},
		{name: "server error logs the cause", apiErr: Internal(), wantLog: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs.Reset()

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products/x", nil)

			Write(c, tt.apiErr, errors.New("mongo is down"))

			if got := logs.Len() > 0; got != tt.wantLog {
				t.Errorf("logged = %v, want %v; logs: %s", got, tt.wantLog, logs.String())
			}
			if recorder.Code != tt.apiErr.StatusCode {
				t.Errorf("status = %d, want %d", recorder.Code, tt.apiErr.StatusCode)
			}
			if body := recorder.Body.String(); !strings.Contains(body, tt.apiErr.Code) {
				t.Errorf("body = %q, want code %q", body, tt.apiErr.Code)
			}
		})
	}
}
