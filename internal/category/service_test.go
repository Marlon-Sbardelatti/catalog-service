package category

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeRepository struct {
	deleted bool
	err     error
}

func (f *fakeRepository) Create(context.Context, *Category) error { return f.err }
func (f *fakeRepository) FindByID(context.Context, bson.ObjectID) (*Category, error) {
	return nil, f.err
}
func (f *fakeRepository) FindAll(context.Context) ([]Category, error) { return nil, f.err }
func (f *fakeRepository) Update(context.Context, bson.ObjectID, bson.M) (*Category, error) {
	return nil, f.err
}
func (f *fakeRepository) Delete(context.Context, bson.ObjectID) error {
	f.deleted = true
	return f.err
}
func (f *fakeRepository) Exists(context.Context, bson.ObjectID) (bool, error) {
	return true, f.err
}

type fakeProductCounter struct {
	count int64
	err   error
}

func (f *fakeProductCounter) CountByCategory(context.Context, bson.ObjectID) (int64, error) {
	return f.count, f.err
}

func TestDelete(t *testing.T) {
	t.Run("blocks deletion when products reference the category", func(t *testing.T) {
		repository := &fakeRepository{}
		service := NewCategoryService(repository, &fakeProductCounter{count: 2})

		if err := service.Delete(context.Background(), bson.NewObjectID()); !errors.Is(err, ErrCategoryInUse) {
			t.Fatalf("Delete() error = %v, want ErrCategoryInUse", err)
		}
		if repository.deleted {
			t.Error("repository.Delete was called, want it skipped")
		}
	})

	t.Run("propagates the product count error", func(t *testing.T) {
		boom := errors.New("mongo is down")
		repository := &fakeRepository{}
		service := NewCategoryService(repository, &fakeProductCounter{err: boom})

		if err := service.Delete(context.Background(), bson.NewObjectID()); !errors.Is(err, boom) {
			t.Fatalf("Delete() error = %v, want %v", err, boom)
		}
		if repository.deleted {
			t.Error("repository.Delete was called, want it skipped")
		}
	})

	t.Run("deletes when no products reference the category", func(t *testing.T) {
		repository := &fakeRepository{}
		service := NewCategoryService(repository, &fakeProductCounter{})

		if err := service.Delete(context.Background(), bson.NewObjectID()); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if !repository.deleted {
			t.Error("repository.Delete was not called")
		}
	})
}
