package category

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	Create(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id bson.ObjectID) (*Category, error)
	FindAll(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, id bson.ObjectID, set bson.M) (*Category, error)
	Delete(ctx context.Context, id bson.ObjectID) error
}
