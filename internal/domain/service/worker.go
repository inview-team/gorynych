package service

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
)

type WorkerService struct {
	accountRepo  entity.AccountRepository
	providerRepo entity.ProviderRepository
	taskQueue    chan entity.ReplicationTask
	resultChan   chan Result
	workerCount  int
}

func NewWorkerService(aRepo entity.AccountRepository, pRepo entity.ProviderRepository, workerCount int) *WorkerService {
	return &WorkerService{
		accountRepo:  aRepo,
		providerRepo: pRepo,
		taskQueue:    make(chan entity.ReplicationTask),
		resultChan:   make(chan Result),
		workerCount:  workerCount,
	}
}

func (s *WorkerService) Start(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		worker := NewReplicationService(i, s.taskQueue, s.resultChan, s.accountRepo, s.providerRepo)
		worker.Start(ctx)
	}
}

func (s *WorkerService) Submit(task entity.ReplicationTask) {
	s.taskQueue <- task
}

func (w *WorkerService) GetResult() Result {
	return <-w.resultChan
}
