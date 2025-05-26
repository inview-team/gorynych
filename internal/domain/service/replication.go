package service

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/s3"
	log "github.com/sirupsen/logrus"
)

const (
	chunkSize int = 100 * 1024 * 1024
)

type ReplicationService struct {
	tasks        <-chan entity.ReplicationTask
	results      chan<- entity.ReplicationResult
	accountRepo  entity.AccountRepository
	providerRepo entity.ProviderRepository
}

type PartTask struct {
	ObjectID       string
	SourceAccount  *entity.ServiceAccount
	SourceProvider *entity.Provider
	TargetAccount  *entity.ServiceAccount
	TargetProvider *entity.Provider
	SourceBucket   string
	TargetBucket   string
	UploadID       string
	PartNumber     int
	Start          int64
	End            int64
}

type PartResult struct {
	PartNumber int
	Tag        string
}

func NewReplicationService(taskQueue <-chan entity.ReplicationTask, resultChan chan<- entity.ReplicationResult, aRepo entity.AccountRepository, pRepo entity.ProviderRepository) *ReplicationService {
	return &ReplicationService{
		tasks:        taskQueue,
		results:      resultChan,
		accountRepo:  aRepo,
		providerRepo: pRepo,
	}
}

func (s *ReplicationService) Start(ctx context.Context) {
	go func() {
		for task := range s.tasks {
			start := time.Now()
			err := s.replicate(ctx, &task)
			end := time.Now()
			s.results <- entity.ReplicationResult{ID: task.ID, Start: start, End: end, Error: err}
		}
	}()
}

func (s *ReplicationService) replicate(ctx context.Context, task *entity.ReplicationTask) error {
	log.Infof("get task to replicate %s from source %s to target %s", task.ObjectID, task.SourceStorage.Bucket, task.TargetStorage.Bucket)
	sourceAccount, sourceProvider, err := s.getAccount(ctx, &task.SourceStorage)
	if err != nil {
		return err
	}

	log.Infof("check existence of  object with id %s", task.ObjectID)
	sourceRepo, err := s3.New(ctx, sourceProvider.Endpoint, sourceAccount.Region, sourceAccount.AccessKey, sourceAccount.Secret)
	object, err := sourceRepo.GetObject(ctx, task.SourceStorage.Bucket, task.ObjectID)
	if err != nil {
		return err
	}

	if object == nil {
		return ErrObjectNotFound
	}

	targetAccount, targetProvider, err := s.getAccount(ctx, &task.TargetStorage)
	if err != nil {
		return err
	}
	targetRepo, err := s3.New(ctx, targetProvider.Endpoint, targetAccount.Region, targetAccount.AccessKey, targetAccount.Secret)
	if err != nil {
		return err
	}

	totalSize := object.Size
	totalParts := int(int64(totalSize)+int64(chunkSize)-1) / chunkSize
	fmt.Print(totalParts)

	uploadID, err := targetRepo.Create(ctx, task.TargetStorage.Bucket, task.ObjectID, nil)
	if err != nil {
		log.Errorf("failed to create upload: %v", err)
		os.Exit(1)
	}

	tasks := make(chan PartTask, totalParts)
	results := make(chan PartResult)

	var wg sync.WaitGroup
	for i := 1; i <= totalParts; i++ {
		wg.Add(1)
		worker := NewWorker(i, tasks, results)
		go worker.Start(ctx, &wg)
	}

	for part := 1; part <= totalParts; part++ {
		start := int64(part-1) * int64(chunkSize)
		end := start + int64(chunkSize) - 1
		if end > totalSize {
			end = totalSize
		}

		tasks <- PartTask{
			ObjectID:       task.ObjectID,
			SourceAccount:  sourceAccount,
			SourceProvider: sourceProvider,
			SourceBucket:   task.SourceStorage.Bucket,
			TargetAccount:  targetAccount,
			TargetProvider: targetProvider,
			TargetBucket:   task.TargetStorage.Bucket,
			UploadID:       uploadID,
			PartNumber:     part,
			Start:          start,
			End:            end,
		}
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	var uploadedParts []entity.UploadPart
	for res := range results {
		uploadedParts = append(uploadedParts, entity.UploadPart{ID: res.Tag, Position: res.PartNumber})
	}

	sort.Slice(uploadedParts, func(i, j int) bool {
		return uploadedParts[i].Position < uploadedParts[j].Position
	})

	err = targetRepo.FinishUpload(ctx, task.TargetStorage.Bucket, uploadID, task.ObjectID, uploadedParts)
	if err != nil {
		log.Errorf("failed to finish upload: %v", err.Error())
	}

	return nil
}

func (s *ReplicationService) getAccount(ctx context.Context, st *entity.Storage) (*entity.ServiceAccount, *entity.Provider, error) {
	log.Info("search bucket")

	account, err := s.accountRepo.GetByID(ctx, st.AccountID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}
	if account == nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	provider, err := s.providerRepo.GetByID(ctx, account.ProviderID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	if provider == nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	oRepo, err := s3.New(ctx, account.ProviderID, account.Region, account.AccessKey, account.Secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	exists, err := oRepo.IsBucketExist(ctx, st.Bucket)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	if !exists {
		return nil, nil, fmt.Errorf("failed to get account: %v", err)
	}

	return account, provider, ErrAccountNotFound
}

type ReplicationWorker struct {
	id      int
	tasks   <-chan PartTask
	results chan<- PartResult
}

func NewWorker(id int, tasks <-chan PartTask, results chan<- PartResult) *ReplicationWorker {
	return &ReplicationWorker{
		id:      id,
		tasks:   tasks,
		results: results,
	}
}

func (w *ReplicationWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range w.tasks {
		sourceRepo, _ := s3.New(ctx, task.SourceProvider.Endpoint, task.SourceAccount.Region, task.SourceAccount.AccessKey, task.SourceAccount.Secret)
		targetRepo, _ := s3.New(ctx, task.TargetProvider.Endpoint, task.TargetAccount.Region, task.TargetAccount.AccessKey, task.TargetAccount.Secret)
		log.Infof("worker%d: part %d: processing bytes %d-%d...", w.id, task.PartNumber, task.Start, task.End)
		reader, err := sourceRepo.StreamDownloadObject(ctx, task.SourceBucket, task.ObjectID, task.Start, task.End)
		if err != nil {
			log.Errorf("failed to download: %v", err.Error())
		}
		partID, err := targetRepo.StreamWritePart(ctx, task.TargetBucket, task.UploadID, task.ObjectID, task.PartNumber, reader)
		if err != nil {
			log.Errorf("failed to download: %v", err.Error())
		}
		w.results <- PartResult{PartNumber: task.PartNumber, Tag: partID}
		log.Infof("worker%d: part %d: DONE", w.id, task.PartNumber)
	}
}
