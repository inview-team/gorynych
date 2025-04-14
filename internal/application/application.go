package application

import (
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
)

type Application struct {
	UploadService      *service.UploadService
	AccountService     *service.AccountService
	ReplicationService *service.ReplicationService
}

func New(client *mongo.Client) *Application {
	aRepo := mongo.NewAccountRepository(client)
	uRepo := mongo.NewUploadRepository(client)
	return &Application{
		service.NewUploadService(uRepo, aRepo),
		service.NewAccountService(aRepo),
		service.NewReplicationService(nil, aRepo),
	}
}
