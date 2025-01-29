package main

import (
	"context"

	"github.com/inview-team/gorynych/internal/application"
	server "github.com/inview-team/gorynych/internal/infrastructure/http"
	"github.com/inview-team/gorynych/pkg/storage/s3/yandex"
)

func main() {
	ctx := context.Background()

	yandexStorage, _ := yandex.New(ctx, "gorynych.1", creds)
	app := application.New()
	app.UploadService.RegisterStorage(ctx, yandexStorage.GetID(), yandexStorage)
	app.ReplicationService.RegisterStorage(ctx, yandexStorage.GetID(), yandexStorage)
	srv := server.NewServer(app)
	srv.Start(ctx)
}
