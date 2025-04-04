package main

import (
	"context"
	"os"

	"github.com/inview-team/gorynych/config"
	"github.com/inview-team/gorynych/internal/application"
	server "github.com/inview-team/gorynych/internal/infrastructure/http"
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

	ctx := context.TODO()
	app := application.New()
	for _, storage := range cfg.Storages {
		switch storage.Provider {
		case "yandex":
			st, err := yandex.New(ctx, storage.AccessKeyID, storage.SecretAccessKey)
			if err != nil {
				log.Error("failed to init storage: %v", err.Error())
			}
			app.UploadService.RegisterStorage(ctx, st)
		default:
			log.Errorf("unknown provider: %s", storage.Provider)
		}

	}

	srv := server.NewServer(app)
	srv.Start(ctx)
}
