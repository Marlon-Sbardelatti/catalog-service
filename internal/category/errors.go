package category

import (
	apierrors "catalog-service/internal/errors"
	"errors"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryUnprocessable = errors.New("invalid category data")
	ErrCategoryInUse         = errors.New("category has products")
)

func toAPIError(err error) apierrors.APIError {
	switch {
	case errors.Is(err, ErrCategoryNotFound):
		return apierrors.NotFound("CATEGORY_NOT_FOUND", "category not found")
	case errors.Is(err, ErrCategoryAlreadyExists):
		return apierrors.Conflict("CATEGORY_ALREADY_EXISTS", "category already exists")
	case errors.Is(err, ErrCategoryUnprocessable):
		return apierrors.UnprocessableEntity("CATEGORY_INVALID", "invalid category data")
	case errors.Is(err, ErrCategoryInUse):
		return apierrors.Conflict("CATEGORY_IN_USE", "category still has products")
	default:
		return apierrors.Internal()
	}
}
