package main

import (
	"context"
	"os"

	"github.com/inview-team/gorynych/config"
	"github.com/inview-team/gorynych/internal/application"
	server "github.com/inview-team/gorynych/internal/infrastructure/http"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
	"github.com/inview-team/gorynych/pkg/storage/s3/yandex"
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

	client, err := mongo.NewClient(ctx, cfg.Storage)
	if err != nil {
		log.Errorf("failed to init database: %v", err.Error())
		os.Exit(1)
	}

	app := application.New(client)

	if len(cfg.Providers) == 0 {
		log.Error("no setup providers")
		os.Exit(1)
	}

	for _, provider := range cfg.Providers {
		switch provider.Type {
		case "yandex":
			st, err := yandex.New(ctx, provider.AccessKeyID, provider.AccessSecret)
			if err != nil {
				log.Error("failed to init storage: %v", err.Error())
			}
			app.UploadService.RegisterStorage(ctx, st)
		default:
			log.Errorf("unknown provider: %s", provider.Type)
		}

	}

	srv := server.NewServer(app)
	srv.Start(ctx)
}
