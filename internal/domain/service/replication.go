package service

import (
	"context"
	"fmt"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/pkg/storage/s3/yandex"
	log "github.com/sirupsen/logrus"
)

type ReplicationService struct {
	rRepo entity.ReplicationRepository
	aRepo entity.AccountRepository
}

func NewReplicationService(rRepo entity.ReplicationRepository, aRepo entity.AccountRepository) *ReplicationService {
	return &ReplicationService{
		rRepo: rRepo,
		aRepo: aRepo,
	}
}

const (
	chunkSize int = 100 * 1024 * 1024
)

func (s *ReplicationService) Replicate(ctx context.Context, objectId string, priority entity.Priority, sourceStorage entity.Storage, targetStorage entity.Storage) error {
	log.Infof("Get task to replicate from source %s to target %s", sourceStorage.Bucket, targetStorage.Bucket)
	sourceRepo, err := s.getAccountByBucket(ctx, sourceStorage)
	if err != nil {
		return err
	}

	log.Infof("check existence of  object with id %s", objectId)
	object, err := sourceRepo.GetObject(ctx, sourceStorage.Bucket, objectId)
	if err != nil {
		return err
	}

	if object == nil {
		return ErrObjectNotFound
	}

	targetRepo, err := s.getAccountByBucket(ctx, targetStorage)
	if err != nil {
		return err
	}

	uploadID, err := targetRepo.Create(ctx, targetStorage.Bucket, objectId, nil)
	if err != nil {
		return fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Create upload for object with ID %s", objectId)
	upload := entity.NewUpload(uploadID, objectId, object.Size, 0, entity.Active, nil, entity.Storage{ProviderID: targetRepo.GetProviderID(ctx), Bucket: targetStorage.Bucket})
	offset := 0

	for {
		log.Infof("Replicate object part from %d to %d", offset, int64(offset)+int64(chunkSize))
		if offset > int(object.Size) {
			break
		}

		data, err := sourceRepo.DownloadObject(ctx, sourceStorage.Bucket, objectId, int64(offset), int64(offset)+int64(chunkSize))
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

	return nil
}

func (s *ReplicationService) getAccountByBucket(ctx context.Context, st entity.Storage) (entity.ObjectRepository, error) {
	log.Info("search bucket")
	accounts, err := s.aRepo.ListByProvider(ctx, entity.Provider(st.ProviderID))
	if err != nil {
		log.Errorf("failed to choose account: failed to list accounts: %v", err.Error())
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}

	for _, account := range accounts {
		oRepo, err := yandex.New(ctx, account.KeyID, account.Secret)
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
