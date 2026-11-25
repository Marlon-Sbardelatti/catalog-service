package app

import (
	"context"
	"fmt"

	"catalog-service/internal/category"
	"catalog-service/internal/config"
	"catalog-service/internal/infrastructure/database/mongodb"
	"catalog-service/internal/infrastructure/server"
	s3storage "catalog-service/internal/infrastructure/storage/s3"
	"catalog-service/internal/product"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router  *gin.Engine
	cleanup func() error
}

func New(cfg *config.Config) (*App, error) {
	db, err := mongodb.NewClientFromConfig(*cfg).Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting to mongodb: %w", err)
	}

	storage, err := s3storage.New(context.Background(), s3storage.Config{
		Region:   cfg.StorageRegion,
		Bucket:   cfg.StorageBucket,
		Endpoint: cfg.StorageEndpoint,
	})
	if err != nil {
		_ = db.Client.Disconnect(context.Background())
		return nil, fmt.Errorf("creating storage client: %w", err)
	}

	categoryRepository := category.NewMongoRepository(db.Database)
	productRepository := product.NewMongoRepository(db.Database)

	return &App{
		Router: server.NewRouter(
			product.NewHandler(
				product.NewProductService(productRepository, storage, categoryRepository),
			),
			category.NewCategoryHandler(
				category.NewCategoryService(categoryRepository, productRepository),
			),
		),
		cleanup: func() error {
			return db.Client.Disconnect(context.Background())
		},
	}, nil
}

func (a *App) Close() error {
	return a.cleanup()
}
