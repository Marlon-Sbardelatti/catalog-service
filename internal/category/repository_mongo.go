package category

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
		collection: db.Collection("categories"),
	}
}

func (r *mongoRepository) Create(
	ctx context.Context,
	category *Category,
) error {
	result, err := r.collection.InsertOne(ctx, category)
	if err != nil {
		return err
	}

	category.ID = result.InsertedID.(bson.ObjectID)

	return nil
}

func (r *mongoRepository) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (*Category, error) {
	var product Category

	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": id},
	).Decode(&product)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrCategoryNotFound
	}

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *mongoRepository) FindAll(ctx context.Context) ([]Category, error) {
	var products []Category

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

func (r *mongoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *mongoRepository) Update(
	ctx context.Context,
	id bson.ObjectID,
	set bson.M,
) (*Category, error) {
	var previous Category

	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.Before),
	).Decode(&previous)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrCategoryNotFound
	}

	if err != nil {
		return nil, err
	}

	return &previous, nil
}
