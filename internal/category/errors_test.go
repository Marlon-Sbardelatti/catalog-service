package category

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestToAPIError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		code       string
	}{
		{
			name:       "wrapped invalid category",
			err:        fmt.Errorf("%w: name cannot be blank", ErrCategoryUnprocessable),
			statusCode: http.StatusUnprocessableEntity,
			code:       "CATEGORY_INVALID",
		},
		{
			name:       "not found",
			err:        ErrCategoryNotFound,
			statusCode: http.StatusNotFound,
			code:       "CATEGORY_NOT_FOUND",
		},
		{
			name:       "already exists",
			err:        ErrCategoryAlreadyExists,
			statusCode: http.StatusConflict,
			code:       "CATEGORY_ALREADY_EXISTS",
		},
		{
			name:       "category in use",
			err:        ErrCategoryInUse,
			statusCode: http.StatusConflict,
			code:       "CATEGORY_IN_USE",
		},
		{
			name:       "unexpected",
			err:        errors.New("mongo is down"),
			statusCode: http.StatusInternalServerError,
			code:       "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := toAPIError(tt.err)

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if apiErr.Code != tt.code {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.code)
			}
		})
	}
}
