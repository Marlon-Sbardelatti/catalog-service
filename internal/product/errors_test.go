package product

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
			name:       "wrapped invalid product",
			err:        fmt.Errorf("%w: name cannot be blank", ErrInvalidProduct),
			statusCode: http.StatusUnprocessableEntity,
			code:       "PRODUCT_INVALID",
		},
		{
			name:       "not found",
			err:        ErrProductNotFound,
			statusCode: http.StatusNotFound,
			code:       "PRODUCT_NOT_FOUND",
		},
		{
			name:       "category not found",
			err:        ErrCategoryNotFound,
			statusCode: http.StatusUnprocessableEntity,
			code:       "PRODUCT_CATEGORY_NOT_FOUND",
		},
		{
			name:       "already exists",
			err:        ErrProductAlreadyExists,
			statusCode: http.StatusConflict,
			code:       "PRODUCT_ALREADY_EXISTS",
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
