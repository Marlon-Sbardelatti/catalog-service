package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	MongoUser     string
	MongoPassword string
	MongoDatabase string
	MongoPort     string
	MongoHost     string

	StorageRegion   string
	StorageBucket   string
	StorageEndpoint string
}

func LoadConfig() (*Config, error) {
	logger := slog.Default()
	if err := godotenv.Load(); err != nil {
		logger.Warn("Could not load .env file", "error", err)
	}

	port := getEnv("PORT", "3001")

	mongoPort := getEnv("MONGO_PORT", "27017")
	mongoHost := getEnv("MONGO_HOST", "localhost")

	mongoUser, err := getRequiredEnv("MONGO_USER")
	if err != nil {
		return nil, err
	}

	mongoPassword, err := getRequiredEnv("MONGO_PASSWORD")
	if err != nil {
		return nil, err
	}

	mongoDatabase, err := getRequiredEnv("MONGO_DATABASE")
	if err != nil {
		return nil, err
	}

	storageBucket, err := getRequiredEnv("S3_BUCKET")
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:          port,
		MongoUser:     mongoUser,
		MongoPassword: mongoPassword,
		MongoDatabase: mongoDatabase,
		MongoPort:     mongoPort,
		MongoHost:     mongoHost,

		StorageRegion:   getEnv("AWS_REGION", "us-east-1"),
		StorageBucket:   storageBucket,
		StorageEndpoint: getEnv("S3_ENDPOINT", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}

func getRequiredEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}

	return value, nil
}
