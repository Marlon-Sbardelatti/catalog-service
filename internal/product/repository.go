package product

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	Create(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id bson.ObjectID) (*Product, error)
	FindAll(ctx context.Context) ([]Product, error)
	Update(ctx context.Context, id bson.ObjectID, set bson.M) (*Product, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	CountByCategory(ctx context.Context, categoryID bson.ObjectID) (int64, error)
}
