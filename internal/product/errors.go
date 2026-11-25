package product

import (
	apierrors "catalog-service/internal/errors"
	"errors"
)

var (
	ErrInvalidProduct       = errors.New("invalid product data")
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
)

func toAPIError(err error) apierrors.APIError {
	switch {
	case errors.Is(err, ErrInvalidProduct):
		return apierrors.UnprocessableEntity("PRODUCT_INVALID", "invalid product data")
	case errors.Is(err, ErrProductNotFound):
		return apierrors.NotFound("PRODUCT_NOT_FOUND", "product not found")
	case errors.Is(err, ErrProductAlreadyExists):
		return apierrors.Conflict("PRODUCT_ALREADY_EXISTS", "product already exists")
	default:
		return apierrors.Internal()
	}
}
