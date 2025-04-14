package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/pkg/storage/s3/yandex"

	log "github.com/sirupsen/logrus"
)

type UploadService struct {
	mu          sync.Mutex
	uploads     map[string]*entity.Upload
	uploadRepo  entity.UploadRepository
	accountRepo entity.AccountRepository
}

func NewUploadService(uRepo entity.UploadRepository, aRepo entity.AccountRepository) *UploadService {
	return &UploadService{
		uploads:     make(map[string]*entity.Upload),
		uploadRepo:  uRepo,
		accountRepo: aRepo,
	}
}

func (s *UploadService) CreateUpload(ctx context.Context, size int64, metadata map[string]string) (string, error) {
	log.Infof("create new upload")
	s.mu.Lock()
	defer s.mu.Unlock()

	oRepo, bucket, err := s.chooseAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %w", err)
	}

	log.Infof("Choose provider: %s and bucket %s", oRepo.GetProviderID(ctx), bucket)

	objectID := entity.NewObjectID()
	uploadID, err := oRepo.Create(ctx, bucket, objectID, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Create upload for object with ID %s", objectID)
	upload := entity.NewUpload(uploadID, objectID, size, 0, entity.Active, nil, entity.Storage{ProviderID: oRepo.GetProviderID(ctx), Bucket: bucket})

	s.uploads[objectID] = upload
	err = s.uploadRepo.Add(ctx, upload)
	if err != nil {
		log.Errorf("failed to save upload: %v", err.Error())
	}

	return objectID, nil
}

func (s *UploadService) chooseAccount(ctx context.Context) (entity.ObjectRepository, string, error) {
	log.Info("choose account for upload")
	accounts, err := s.accountRepo.ListByProvider(ctx, entity.Yandex)
	if err != nil {
		log.Errorf("failed to choose account: failed to list accounts: %v", err.Error())
		return nil, "", err
	}

	if len(accounts) == 0 {
		return nil, "", ErrNoAvailableAccounts
	}

	for _, account := range accounts {
		oRepo, err := yandex.New(ctx, account.KeyID, account.Secret)
		if err != nil {
			log.Errorf("failed to init storage by account with id: %s", account.ID)
			continue
		}

		buckets, err := oRepo.ListBuckets(ctx)
		if err != nil {
			log.Errorf("failed to get bucket: %v", err)
			continue
		}
		log.Infof("found %d buckets: %v", len(buckets), buckets)

		return oRepo, buckets[0], nil
	}
	return nil, "", ErrNoAvailableBuckets
}

func (s *UploadService) WritePart(ctx context.Context, objectID string, offset int64, data []byte) (int64, error) {
	log.Infof("write part to object with id %s", objectID)
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Infof("Search for upload with Object ID %s", objectID)
	var upload *entity.Upload
	upload, exists := s.uploads[objectID]
	if !exists {
		upload, err := s.uploadRepo.GetByID(ctx, objectID)
		if err != nil {
			log.Errorf("failed to write part: failed to find upload: %v", err.Error())
			return 0, err
		}

		if upload == nil {
			return 0, ErrUploadNotFound
		}
	}

	log.Infof("Update upload: %v\n", *upload)
	if offset != upload.Offset {
		return 0, ErrWrongOffset
	}

	if offset+int64(len(data)) > upload.Size {
		return 0, ErrUploadBig
	}

	log.Infof("Search account from Provider %s with access to bucket %s", upload.Storage.ProviderID, upload.Storage.Bucket)
	oRepo, err := s.getAccountByBucket(ctx, upload.Storage)
	if err != nil {
		return 0, ErrBucketNotFound
	}

	var position int
	if upload.Parts != nil {
		position = len(upload.Parts) + 1
	} else {
		position = 1
	}

	partID, err := oRepo.WritePart(ctx, upload.Storage.Bucket, upload.ID, upload.ObjectID, position, data)
	if err != nil {
		log.Errorf("failed to write part. Reason: %v", err)
		return 0, fmt.Errorf("failed to upload chunk: %w", err)
	}

	upload.SetOffset(offset + int64(len(data)))
	upload.AddPartial(partID, position)

	if upload.Offset == upload.Size {
		err := oRepo.FinishUpload(ctx, upload)
		if err != nil {
			return 0, fmt.Errorf("failed to finish upload: %v", err)
		}
		upload.Status = entity.Complete
		delete(s.uploads, objectID)
	}

	err = s.uploadRepo.Update(ctx, upload)
	if err != nil {
		log.Errorf("failed to update upload: %v", err)
	}

	return upload.Offset, nil
}

func (s *UploadService) getAccountByBucket(ctx context.Context, st entity.Storage) (entity.ObjectRepository, error) {
	log.Info("search bucket")
	accounts, err := s.accountRepo.ListByProvider(ctx, entity.Provider(st.ProviderID))
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

func (s *UploadService) GetUploads() []*entity.Upload {
	s.mu.Lock()
	defer s.mu.Unlock()
	var uploads []*entity.Upload
	for _, upload := range s.uploads {
		uploads = append(uploads, upload)
	}
	return uploads
}

func (s *UploadService) GetUpload(ctx context.Context, id string) (*entity.Upload, error) {
	upload, exists := s.uploads[id]
	if !exists {
		return nil, ErrUploadNotFound
	}
	return upload, nil
}
