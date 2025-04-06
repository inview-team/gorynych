package mongo

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	Database *mongo.Database
}

type Config struct {
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
}

var (
	DefaultConfig = Config{
		Host:     "localhost",
		Port:     27017,
		Database: "gorynych",
	}
)

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Errorf("failed to init client: %v", err.Error())
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &Client{
		Database: client.Database(cfg.Database),
	}, nil
}
