package service

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/inview-team/gorynych/internal/domain/entity"
)

type TaskService struct {
	accountRepo  entity.AccountRepository
	providerRepo entity.ProviderRepository
	taskRepo     entity.TaskRepository
	tasksChan    chan entity.ReplicationTask
	resultChan   chan entity.ReplicationResult
	workerCount  int
}

func NewTaskService(aRepo entity.AccountRepository, pRepo entity.ProviderRepository, tRepo entity.TaskRepository, workerCount int) *TaskService {
	return &TaskService{
		accountRepo:  aRepo,
		providerRepo: pRepo,
		taskRepo:     tRepo,
		tasksChan:    make(chan entity.ReplicationTask),
		resultChan:   make(chan entity.ReplicationResult),
		workerCount:  workerCount,
	}
}

func (s *TaskService) Start(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		worker := NewReplicationService(s.tasksChan, s.resultChan, s.accountRepo, s.providerRepo)
		worker.Start(ctx)
	}

	go func() {
		for result := range s.resultChan {
			task, err := s.taskRepo.GetByID(ctx, result.ID)
			if err != nil {
				log.Errorf("failed to save result of task %s. Reason: %v", result.ID, err)
			}
			if result.Error != nil {
				log.Errorf("task %s failed. Reason: %v", result.ID, result.Error)
				task.Status = entity.TaskFailed
			} else {
				task.Status = entity.TaskCompleted
			}
			task.Start = result.Start
			task.End = result.End
			err = s.taskRepo.Update(ctx, task)
			if err != nil {
				log.Errorf("failed to save result of task %s. Reason: %v", result.ID, err)
			}
		}
	}()
}

func (s *TaskService) Replication(ctx context.Context, objectID string, sourceStorage, targetStorage entity.Storage) (string, error) {
	task := entity.ReplicationTask{
		ID:            entity.NewTaskID(),
		ObjectID:      objectID,
		SourceStorage: sourceStorage,
		TargetStorage: targetStorage,
	}
	err := s.taskRepo.Add(ctx, &entity.Task{ID: task.ID, Start: time.Time{}, End: time.Time{}, Type: entity.Replication, Status: entity.TaskCreated})
	if err != nil {
		log.Errorf("failed to create replication task: %v", err)
		return "", err
	}
	s.tasksChan <- task
	return task.ID, err
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (*entity.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		log.Errorf("failed to get result of task %s. Reason: %v", taskID, err)
		return nil, err
	}

	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, err
}
