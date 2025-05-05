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

func New(ctx context.Context, client *mongo.Client) (*Application, error) {
	aRepo := mongo.NewAccountRepository(client)
	uRepo := mongo.NewUploadRepository(client)
	pRepo, err := mongo.NewProviderRepository(ctx, client)
	if err != nil {
		return nil, err
	}
	workerService := service.NewWorkerService(aRepo, pRepo, 5)
	workerService.Start(ctx)
	return &Application{
		service.NewUploadService(uRepo, aRepo, pRepo),
		service.NewAccountService(aRepo),
		workerService,
	}, nil
}
