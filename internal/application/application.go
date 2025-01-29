package application

import (
	"github.com/inview-team/gorynych/internal/domain/service"
)

type Application struct {
	UploadService      *service.UploadService
	ReplicationService *service.ReplicationService
}

func New() *Application {
	return &Application{
		service.New(service.Config{MaxSize: 0}),
		service.NewReplicationService(),
	}

}
