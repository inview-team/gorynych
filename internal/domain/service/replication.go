package service

import (
	"context"
	"fmt"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/s3"
	log "github.com/sirupsen/logrus"
)

type ReplicationWorker struct {
	id         int
	aRepo      entity.AccountRepository
	pRepo      entity.ProviderRepository
	taskQueue  <-chan entity.ReplicationTask
	resultChan chan<- Result
}

type Result struct {
	ObjectID string
	Storage  entity.Storage
	Error    error
}

func NewReplicationService(id int, taskQueue <-chan entity.ReplicationTask, resultChan chan<- Result, aRepo entity.AccountRepository, pRepo entity.ProviderRepository) *ReplicationWorker {
	return &ReplicationWorker{
		id:         id,
		taskQueue:  taskQueue,
		resultChan: resultChan,
		aRepo:      aRepo,
		pRepo:      pRepo,
	}
}

const (
	chunkSize int = 100 * 1024 * 1024
)

func (w *ReplicationWorker) Start(ctx context.Context) {
	go func() {
		for task := range w.taskQueue {
			err := w.replicate(ctx, &task)
			if err != nil {
				w.resultChan <- Result{ObjectID: task.ObjectID, Storage: entity.Storage{}, Error: err}
				continue
			}
			w.resultChan <- Result{ObjectID: task.ObjectID, Storage: task.TargetStorage, Error: nil}
		}
	}()
}

func (w *ReplicationWorker) replicate(ctx context.Context, task *entity.ReplicationTask) error {
	log.Infof("Worker with id %d receive replication task", w.id)
	log.Infof("Get task to replicate from source %s to target %s", task.SourceStorage.Bucket, task.TargetStorage.Bucket)
	sourceRepo, err := w.getAccountByBucket(ctx, task.SourceStorage)
	if err != nil {
		return err
	}

	log.Infof("check existence of  object with id %s", task.ObjectID)
	object, err := sourceRepo.GetObject(ctx, task.SourceStorage.Bucket, task.ObjectID)
	if err != nil {
		return err
	}

	if object == nil {
		return ErrObjectNotFound
	}

	targetRepo, err := w.getAccountByBucket(ctx, task.TargetStorage)
	if err != nil {
		return err
	}

	uploadID, err := targetRepo.Create(ctx, task.TargetStorage.Bucket, task.ObjectID, nil)
	if err != nil {
		return fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Create upload for object with ID %s", task.ObjectID)
	upload := entity.NewUpload(uploadID, task.ObjectID, object.Size, 0, entity.Active, nil, entity.Storage{ProviderID: task.TargetStorage.ProviderID, Bucket: task.TargetStorage.Bucket})
	offset := 0

	for {
		log.Infof("Replicate object part from %d to %d", offset, int64(offset)+int64(chunkSize))
		if offset > int(object.Size) {
			break
		}

		data, err := sourceRepo.DownloadObject(ctx, task.SourceStorage.Bucket, task.ObjectID, int64(offset), int64(offset)+int64(chunkSize))
		if err != nil {
			return err
		}

		log.Infof("Download part of file %d", len(data))

		var position int
		if upload.Parts != nil {
			position = len(upload.Parts) + 1
		} else {
			position = 1
		}

		partID, err := targetRepo.WritePart(ctx, upload.Storage.Bucket, upload.ID, upload.ObjectID, position, data)
		if err != nil {
			log.Errorf("failed to write part. Reason: %v", err)
			return fmt.Errorf("failed to upload chunk: %w", err)
		}
		log.Infof("Upload chunk with id %s", partID)

		offset += chunkSize
		upload.AddPartial(partID, position)
	}

	err = targetRepo.FinishUpload(ctx, upload)
	if err != nil {
		return fmt.Errorf("failed to finish upload: %v", err)
	}
	log.Info("Replication process done")
	return nil
}

func (w *ReplicationWorker) getAccountByBucket(ctx context.Context, st entity.Storage) (entity.ObjectRepository, error) {
	log.Info("search bucket")
	provider, err := w.pRepo.GetByID(ctx, st.ProviderID)
	if err != nil {
		log.Errorf("failed to choose account: failed to find provider: %v", err.Error())
		return nil, err
	}
	accounts, err := w.aRepo.ListByProvider(ctx, st.ProviderID)

	if err != nil {
		log.Errorf("failed to choose account: failed to list accounts: %v", err.Error())
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	for _, account := range accounts {
		oRepo, err := s3.New(ctx, provider.Endpoint, account.Region, account.AccessKey, account.Secret)
		if err != nil {
			log.Errorf("failed to init storage by account with id: %s", account.ID)
			continue
		}

		exists, err := oRepo.IsBucketExist(ctx, st.Bucket)
		if err != nil {
			continue
		}

		if !exists {
			continue
		}
		return oRepo, nil
	}

	return nil, ErrNoAvailableBuckets
}
