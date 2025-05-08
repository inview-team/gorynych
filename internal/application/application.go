package application

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
)

type Application struct {
	UploadService  *service.UploadService
	AccountService *service.AccountService
	TaskService    *service.TaskService
}

func New(ctx context.Context, client *mongo.Client) (*Application, error) {
	aRepo := mongo.NewAccountRepository(client)
	uRepo := mongo.NewUploadRepository(client)
	pRepo, err := mongo.NewProviderRepository(ctx, client)
	tRepo := mongo.NewTaskRepository(client)
	if err != nil {
		return nil, err
	}
	taskService := service.NewTaskService(aRepo, pRepo, tRepo, 5)
	taskService.Start(ctx)
	return &Application{
		service.NewUploadService(uRepo, aRepo, pRepo),
		service.NewAccountService(aRepo),
		taskService,
	}, nil
}
