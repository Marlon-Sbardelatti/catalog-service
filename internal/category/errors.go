package category

import (
	apierrors "catalog-service/internal/errors"
	"errors"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryUnprocessable = errors.New("invalid category data")
)

func toAPIError(err error) apierrors.APIError {
	switch {
	case errors.Is(err, ErrCategoryNotFound):
		return apierrors.NotFound("CATEGORY_NOT_FOUND", "category not found")
	case errors.Is(err, ErrCategoryAlreadyExists):
		return apierrors.Conflict("CATEGORY_ALREADY_EXISTS", "category already exists")
	case errors.Is(err, ErrCategoryUnprocessable):
		return apierrors.UnprocessableEntity("CATEGORY_INVALID", "invalid category data")
	default:
		return apierrors.Internal()
	}
}
