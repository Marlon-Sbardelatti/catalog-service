package product

import (
	"catalog-service/internal/utils"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FileStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type CategoryChecker interface {
	Exists(ctx context.Context, id bson.ObjectID) (bool, error)
}

type ProductService struct {
	repository Repository
	storage    FileStorage
	categories CategoryChecker
}

func NewProductService(repository Repository, storage FileStorage, categories CategoryChecker) *ProductService {
	return &ProductService{
		repository: repository,
		storage:    storage,
		categories: categories,
	}
}

func (s *ProductService) Create(
	ctx context.Context,
	req CreateProductRequest,
	categoryID bson.ObjectID,
	file *multipart.FileHeader,
) (*Product, error) {
	product := &Product{
		CategoryID:  categoryID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       req.Price,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if product.Name == "" || product.Price <= 0 {
		return nil, ErrInvalidProduct
	}

	if err := s.ensureCategoryExists(ctx, categoryID); err != nil {
		return nil, err
	}

	contentType, src, err := utils.ReadImage(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidProduct, err)
	}
	defer src.Close()

	product.ImageKey = "products/" + uuid.NewString()

	if err := s.storage.Upload(ctx, product.ImageKey, src, file.Size, contentType); err != nil {
		return nil, fmt.Errorf("uploading image: %w", err)
	}

	if err := s.repository.Create(ctx, product); err != nil {
		if deleteErr := s.storage.Delete(ctx, product.ImageKey); deleteErr != nil {
			slog.Error("deleting unused product image", "key", product.ImageKey, "error", deleteErr)
		}
		return nil, err
	}

	product.ImageURL = s.storage.PublicURL(product.ImageKey)

	return product, nil
}

func (s *ProductService) FindAll(ctx context.Context) ([]Product, error) {
	products, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].ImageURL = s.storage.PublicURL(products[i].ImageKey)
	}

	return products, nil
}

func (s *ProductService) FindByID(ctx context.Context, id bson.ObjectID) (*Product, error) {
	product, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	product.ImageURL = s.storage.PublicURL(product.ImageKey)
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id bson.ObjectID) error {
	return s.repository.Delete(ctx, id)
}

func (s *ProductService) Update(ctx context.Context,
	id bson.ObjectID,
	req UpdateProductRequest,
	categoryID bson.ObjectID,
	file *multipart.FileHeader,
) (*Product, error) {
	set := bson.M{}

	var (
		name        string
		description string
		price       int64
	)

	if req.Name != "" {
		name = strings.TrimSpace(req.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be blank", ErrInvalidProduct)
		}
		set["name"] = name
	}

	if req.Description != "" {
		description = strings.TrimSpace(req.Description)
		set["description"] = description
	}

	if req.Price > 0 {
		price = req.Price
		set["price"] = price
	}

	if !categoryID.IsZero() {
		if err := s.ensureCategoryExists(ctx, categoryID); err != nil {
			return nil, err
		}
		set["category_id"] = categoryID
	}

	if len(set) == 0 && file == nil {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidProduct)
	}

	var newKey string

	if file != nil {
		contentType, src, err := utils.ReadImage(file)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidProduct, err)
		}
		defer src.Close()

		newKey = "products/" + uuid.NewString()

		if err := s.storage.Upload(ctx, newKey, src, file.Size, contentType); err != nil {
			return nil, fmt.Errorf("uploading image: %w", err)
		}

		set["image_key"] = newKey
	}

	now := time.Now()
	set["updated_at"] = now

	previous, err := s.repository.Update(ctx, id, set)
	if err != nil {
		if newKey != "" {
			if deleteErr := s.storage.Delete(ctx, newKey); deleteErr != nil {
				slog.Error("deleting unused product image", "key", newKey, "error", deleteErr)
			}
		}
		return nil, err
	}

	if newKey != "" && previous.ImageKey != "" && previous.ImageKey != newKey {
		if err := s.storage.Delete(ctx, previous.ImageKey); err != nil {
			slog.Error("deleting previous product image", "key", previous.ImageKey, "error", err)
		}
	}

	product := *previous

	if name != "" {
		product.Name = name
	}
	if description != "" {
		product.Description = description
	}
	if price > 0 {
		product.Price = price
	}
	if !categoryID.IsZero() {
		product.CategoryID = categoryID
	}
	if newKey != "" {
		product.ImageKey = newKey
	}
	product.UpdatedAt = now
	product.ImageURL = s.storage.PublicURL(product.ImageKey)

	return &product, nil
}

func (s *ProductService) ensureCategoryExists(ctx context.Context, id bson.ObjectID) error {
	exists, err := s.categories.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("checking category: %w", err)
	}

	if !exists {
		return ErrCategoryNotFound
	}

	return nil
}
