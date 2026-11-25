package product

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoRepository struct {
	collection *mongo.Collection
}

var _ Repository = (*mongoRepository)(nil)

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("products"),
	}
}

func (r *mongoRepository) Create(
	ctx context.Context,
	product *Product,
) error {
	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return err
	}

	product.ID = result.InsertedID.(bson.ObjectID)

	return nil
}

func (r *mongoRepository) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (*Product, error) {
	var product Product

	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": id},
	).Decode(&product)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProductNotFound
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *mongoRepository) FindAll(ctx context.Context) ([]Product, error) {
	var products []Product

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *mongoRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	set bson.M,
) (*Product, error) {
	var previous Product

	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.Before),
	).Decode(&previous)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProductNotFound
	}

	if err != nil {
		return nil, err
	}

	return &previous, nil
}

func (r *mongoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrProductNotFound
	}
	return nil
}
