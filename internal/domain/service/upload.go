package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/s3"

	log "github.com/sirupsen/logrus"
)

type UploadService struct {
	mu           sync.Mutex
	uploads      map[string]*entity.Upload
	uploadRepo   entity.UploadRepository
	providerRepo entity.ProviderRepository
	accountRepo  entity.AccountRepository
}

func NewUploadService(uRepo entity.UploadRepository, aRepo entity.AccountRepository, pRepo entity.ProviderRepository) *UploadService {
	return &UploadService{
		uploads:      make(map[string]*entity.Upload),
		uploadRepo:   uRepo,
		accountRepo:  aRepo,
		providerRepo: pRepo,
	}
}

func (s *UploadService) CreateUpload(ctx context.Context, size int64, metadata map[string]string, accountID string, bucket string) (string, error) {
	log.Infof("create new upload")
	s.mu.Lock()
	defer s.mu.Unlock()

	storage := entity.Storage{AccountID: accountID, Bucket: bucket}
	account, provider, err := s.getAccount(ctx, &storage)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	oRepo, err := s3.New(ctx, provider.Endpoint, account.Region, account.AccessKey, account.Secret)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	bucketExists, err := oRepo.IsBucketExist(ctx, bucket)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	if !bucketExists {
		return "", ErrBucketNotFound
	}

	objectID := entity.NewObjectID()
	uploadID, err := oRepo.Create(ctx, bucket, objectID, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Create upload for object with ID %s", objectID)
	upload := entity.NewUpload(uploadID, objectID, size, 0, entity.Active, nil, storage)

	s.uploads[objectID] = upload
	err = s.uploadRepo.Add(ctx, upload)
	if err != nil {
		log.Errorf("failed to save upload: %v", err.Error())
	}

	return objectID, nil
}

func (s *UploadService) WritePart(ctx context.Context, objectID string, offset int64, data *[]byte) (int64, error) {
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

	account, provider, err := s.getAccount(ctx, &upload.Storage)
	if err != nil {
		return 0, fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Update upload: %v\n", *upload)
	if offset != upload.Offset {
		return 0, ErrWrongOffset
	}

	if offset+int64(len(*data)) > upload.Size {
		return 0, ErrUploadBig
	}

	oRepo, err := s3.New(ctx, provider.Endpoint, account.Region, account.AccessKey, account.Secret)
	if err != nil {
		return 0, fmt.Errorf("failed to create upload: %v", err)
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

	upload.SetOffset(offset + int64(len(*data)))
	upload.AddPartial(partID, position)

	if upload.Offset == upload.Size {
		err := oRepo.FinishUpload(ctx, upload.Storage.Bucket, upload.ID, upload.ObjectID, upload.Parts)
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

func (s *UploadService) getAccount(ctx context.Context, st *entity.Storage) (*entity.ServiceAccount, *entity.Provider, error) {
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
