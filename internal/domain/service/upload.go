package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/inview-team/gorynych/internal/domain/entity"
)

type UploadService struct {
	mu       sync.Mutex
	storages map[string]entity.UploadRepository
	uploads  map[entity.UploadID]*entity.Upload
	maxSize  int64
}

func New(maxSize int64) *UploadService {
	return &UploadService{
		storages: make(map[string]entity.UploadRepository),
		uploads:  make(map[entity.UploadID]*entity.Upload),
		maxSize:  maxSize,
	}
}

func (s *UploadService) RegisterStorage(ctx context.Context, id string, storage entity.UploadRepository) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storages[id]; ok {
		return ErrRepositoryExists
	}
	s.storages[id] = storage
	return nil
}

func (s *UploadService) DeregisterStorage(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storages[id]; !ok {
		return ErrRepositoryNotFound
	}
	delete(s.storages, id)
	return nil
}

func (s *UploadService) CreateUpload(ctx context.Context, size int64, metadata map[string]string) (entity.UploadID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if size != 0 && s.maxSize > 0 && size > s.maxSize {
		return "", ErrResourceTooBig
	}

	upload := entity.NewUpload(size, metadata)
	fmt.Println(*upload)

	id, err := s.chooseUploadStorage()
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}
	fmt.Println(id)

	err = s.storages[id].Create(ctx, upload)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	upload.SetStorage(entity.StorageID(id))
	s.uploads[upload.ID] = upload
	fmt.Println(*upload)
	return upload.ID, nil
}

func (s *UploadService) WritePart(ctx context.Context, id entity.UploadID, offset int64, data []byte) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	upload, exists := s.uploads[id]
	if !exists {
		return 0, errors.New("upload not found")
	}

	if offset != upload.Offset {
		return 0, fmt.Errorf("incorrect offset provided. Expected: %d, Got: %d", upload.Offset, offset)
	}

	if offset+int64(len(data)) > upload.Size {
		return 0, fmt.Errorf("chunk exceeds file size. File size: %d, chunk offset: %d, chunk size: %d", upload.Size, offset, len(data))
	}

	storage, exists := s.storages[string(upload.StorageID)]
	if !exists {
		return 0, errors.New("storage not found")
	}

	err := storage.WriteChunk(ctx, upload.ID, offset, data)
	if err != nil {
		return 0, fmt.Errorf("failed to upload chunk: %v", err)
	}

	upload.SetOffset(offset + int64(len(data)))
	if upload.Offset == upload.Size {
		upload.Status = entity.Complete
		err := storage.FinishUpload(ctx, upload.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to finish upload: %v", err)
		}
		return -1, nil
	}
	return upload.Offset, err
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

func (s *UploadService) GetStorages() []entity.UploadRepository {
	s.mu.Lock()
	defer s.mu.Unlock()
	var storages []entity.UploadRepository
	for _, storage := range s.storages {
		storages = append(storages, storage)
	}
	return storages
}

func (s *UploadService) chooseUploadStorage() (string, error) {
	if len(s.storages) == 0 {
		return "", errors.New("no available repositories")
	}

	for id := range s.storages {
		return id, nil
	}
	return "", ErrNoAvailableBuckets
}
