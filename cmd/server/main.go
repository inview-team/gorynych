package main

import (
	"context"
	"os"

	"github.com/inview-team/gorynych/config"
	"github.com/inview-team/gorynych/internal/application"
	server "github.com/inview-team/gorynych/internal/infrastructure/http"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
	log "github.com/sirupsen/logrus"
)

func main() {
	configPath := os.Getenv("SERVICE_CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	cfg, err := config.LoadFile(configPath)
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	log.Info(cfg)

	ctx := context.TODO()

	client, err := mongo.NewClient(ctx, cfg.Database)
	if err != nil {
		log.Errorf("failed to init database: %v", err.Error())
		os.Exit(1)
	}

	app, err := application.New(ctx, client)
	if err != nil {
		log.Errorf("failed to init application: %v", err.Error())
		os.Exit(1)
	}

	srv := server.NewServer(app)
	srv.Start(ctx)
}
