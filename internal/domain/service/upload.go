package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/inview-team/gorynych/internal/domain/entity"

	log "github.com/sirupsen/logrus"
)

type UploadService struct {
	mu         sync.Mutex
	storages   map[string]entity.StorageRepository
	uploads    map[string]*entity.Upload
	uploadRepo entity.UploadRepository
}

func NewUploadService(uRepo entity.UploadRepository) *UploadService {
	return &UploadService{
		storages:   make(map[string]entity.StorageRepository),
		uploads:    make(map[string]*entity.Upload),
		uploadRepo: uRepo,
	}
}

func (s *UploadService) RegisterStorage(ctx context.Context, storage entity.StorageRepository) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	storageId := entity.NewStorageID()
	if _, ok := s.storages[storageId]; ok {
		return ErrStorageExists
	}
	log.Infof("registered storage with id %s from provider %s", storageId, storage.GetProviderID(ctx))
	s.storages[storageId] = storage
	return nil
}

func (s *UploadService) DeregisterStorage(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storages[id]; !ok {
		return ErrStorageNotFound
	}
	delete(s.storages, id)
	return nil
}

func (s *UploadService) CreateUpload(ctx context.Context, size int64, metadata map[string]string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageID, bucket, err := s.chooseBucket(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %w", err)
	}

	log.Infof("Choose storage with id: %s and bucket %s", storageID, bucket)
	storage := s.storages[storageID]
	objectID := entity.NewObjectID()
	uploadID, err := storage.Create(ctx, bucket, objectID, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	log.Infof("Create upload for object with ID %s", objectID)
	upload := entity.NewUpload(uploadID, objectID, size, 0, entity.Active, nil, entity.Storage{ProviderID: storage.GetProviderID(ctx), Bucket: bucket})
	s.uploads[objectID] = upload
	err = s.uploadRepo.Add(ctx, upload)
	if err != nil {
		log.Errorf("failed to save upload: %v", err.Error())
	}

	return objectID, nil
}

func (s *UploadService) chooseBucket(ctx context.Context) (string, string, error) {
	if len(s.storages) == 0 {
		return "", "", errors.New("no available storages")
	}

	for id, storage := range s.storages {
		buckets, err := storage.ListBuckets(ctx)
		if err != nil {
			log.Errorf("failed to get bucket: %v", err)
			continue
		}
		log.Infof("found %d buckets: %v", len(buckets), buckets)

		return id, buckets[0], nil
	}
	return "", "", ErrNoAvailableBuckets
}

func (s *UploadService) WritePart(ctx context.Context, objectID string, offset int64, data []byte) (int64, error) {
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

	log.Infof("Search for storage with Bucket %s", upload.Storage.Bucket)
	storageID, err := s.getStorageByBucket(ctx, upload.Storage)
	if err != nil {
		return 0, ErrBucketNotFound
	}
	log.Infof("Found storage with id %s", storageID)
	storage := s.storages[storageID]

	var position int
	if upload.Parts != nil {
		position = len(upload.Parts) + 1
	} else {
		position = 1
	}

	partID, err := storage.WritePart(ctx, upload.Storage.Bucket, upload.ID, upload.ObjectID, position, data)
	if err != nil {
		log.Errorf("failed to write part. Reason: %v", err)
		return 0, fmt.Errorf("failed to upload chunk: %w", err)
	}

	upload.SetOffset(offset + int64(len(data)))
	upload.AddPartial(partID, position)

	if upload.Offset == upload.Size {
		err := storage.FinishUpload(ctx, upload)
		if err != nil {
			return 0, fmt.Errorf("failed to finish upload: %v", err)
		}
		upload.Status = entity.Complete
	}

	err = s.uploadRepo.Update(ctx, upload)
	if err != nil {
		log.Errorf("failed to update upload: %v", err)
	}

	return upload.Offset, nil
}

func (s *UploadService) getStorageByBucket(ctx context.Context, st entity.Storage) (string, error) {
	if len(s.storages) == 0 {
		return "", errors.New("no available storages")
	}

	for id, storage := range s.storages {
		if st.ProviderID != storage.GetProviderID(ctx) {
			continue
		}

		exists, err := storage.IsBucketExist(ctx, st.Bucket)
		if err != nil {
			// fmt.Errorf("failed to retrieve buckets: %v")
			continue
		}

		if !exists {
			continue
		}
		return id, nil
	}
	return "", ErrNoAvailableBuckets
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
