package application

import (
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
)

type Application struct {
	UploadService *service.UploadService
}

func New(client *mongo.Client) *Application {
	return &Application{
		service.NewUploadService(mongo.NewUploadRepository(client)),
	}
}
