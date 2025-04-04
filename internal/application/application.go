package application

import (
	"github.com/inview-team/gorynych/internal/domain/service"
)

type Application struct {
	UploadService *service.UploadService
}

func New() *Application {
	return &Application{
		service.NewUploadService(nil),
	}
}
