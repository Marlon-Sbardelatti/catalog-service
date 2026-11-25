package mongodb

import (
	"catalog-service/internal/config"
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func NewClientFromConfig(cfg config.Config) *Client {
	return &Client{
		Host:     cfg.MongoHost,
		Port:     cfg.MongoPort,
		Username: cfg.MongoUser,
		Password: cfg.MongoPassword,
		Database: cfg.MongoDatabase,
	}
}

func (c *Client) GenerateURI() string {
	uri := url.URL{
		Scheme: "mongodb",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/", // mongo requires a / before the query string
		RawQuery: url.Values{
			"authSource": {"admin"},
		}.Encode(),
	}

	return uri.String()
}

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func (c *Client) Connect() (*MongoDB, error) {
	uri := c.GenerateURI()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("pinging mongodb: %w", err)
	}

	return &MongoDB{
		Client:   client,
		Database: client.Database(c.Database),
	}, nil
}
