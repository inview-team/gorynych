package application

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
)

type Application struct {
	UploadService  *service.UploadService
	AccountService *service.AccountService
	WorkerService  *service.WorkerService
}

func New(ctx context.Context, client *mongo.Client) *Application {
	aRepo := mongo.NewAccountRepository(client)
	uRepo := mongo.NewUploadRepository(client)
	workerService := service.NewWorkerService(aRepo, 5)
	workerService.Start(ctx)
	return &Application{
		service.NewUploadService(uRepo, aRepo),
		service.NewAccountService(aRepo),
		workerService,
	}
}
