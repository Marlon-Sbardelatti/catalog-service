package category

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductCounter interface {
	CountByCategory(ctx context.Context, categoryID bson.ObjectID) (int64, error)
}

type CategoryService struct {
	repository Repository
	products   ProductCounter
}

func NewCategoryService(repository Repository, products ProductCounter) *CategoryService {
	return &CategoryService{
		repository: repository,
		products:   products,
	}
}

func (s *CategoryService) Create(ctx context.Context, req CreateCategoryRequest) (*Category, error) {
	category := &Category{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if category.Name == "" {
		return nil, ErrCategoryUnprocessable
	}

	if err := s.repository.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) FindAll(ctx context.Context) ([]Category, error) {
	categories, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *CategoryService) FindByID(ctx context.Context, id bson.ObjectID) (*Category, error) {
	category, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id bson.ObjectID) error {
	count, err := s.products.CountByCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("counting products: %w", err)
	}

	if count > 0 {
		return ErrCategoryInUse
	}

	return s.repository.Delete(ctx, id)
}

func (s *CategoryService) Update(ctx context.Context, id bson.ObjectID, req UpdateCategoryRequest) (*Category, error) {
	set := bson.M{}

	var (
		name        string
		description string
	)

	if req.Name != "" {
		name = strings.TrimSpace(req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be blank", ErrCategoryUnprocessable)
		}
		set["name"] = name
	}

	if req.Description != "" {
		description = strings.TrimSpace(req.Description)
		set["description"] = description
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("%w: no fields to update", ErrCategoryUnprocessable)
	}

	now := time.Now()
	set["updated_at"] = now

	previous, err := s.repository.Update(ctx, id, set)
	if err != nil {
		return nil, err
	}

	category := *previous

	if name != "" {
		category.Name = name
	}
	if description != "" {
		category.Description = description
	}
	category.UpdatedAt = now

	return &category, nil
}
