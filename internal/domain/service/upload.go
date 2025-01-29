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
	config   Config
}

type Config struct {
	MaxSize int64
}

func New(c Config) *UploadService {
	return &UploadService{
		storages: make(map[string]entity.UploadRepository),
		uploads:  make(map[entity.UploadID]*entity.Upload),
		config:   c,
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

	if size != 0 && s.config.MaxSize > 0 && size > s.config.MaxSize {
		return "", ErrUploadBig
	}

	upload := entity.NewUpload(size, metadata)

	id, err := s.chooseUploadStorage()
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	err = s.storages[id].Create(ctx, upload)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	upload.SetStorage(entity.StorageID(id))
	s.uploads[upload.ID] = upload
	fmt.Printf("Create upload: %v\n", *upload)
	return upload.ID, nil
}

func (s *UploadService) WritePart(ctx context.Context, id entity.UploadID, offset int64, data []byte) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	upload, exists := s.uploads[id]
	if !exists {
		return 0, ErrUploadNotFound
	}
	fmt.Printf("Update upload: %v\n", *upload)
	if offset != upload.Offset {
		return 0, ErrWrongOffset
	}

	if offset+int64(len(data)) > upload.Size {
		return 0, ErrUploadBig
	}

	storage, exists := s.storages[string(upload.StorageID)]
	if !exists {
		return 0, ErrRepositoryNotFound
	}

	err := storage.WriteChunk(ctx, upload.ID, offset, data)
	if err != nil {
		return 0, fmt.Errorf("failed to upload chunk: %v", err)
	}

	upload.SetOffset(offset + int64(len(data)))
	if upload.Offset == upload.Size {
		err := storage.FinishUpload(ctx, upload.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to finish upload: %v", err)
		}
		upload.Status = entity.Complete
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

func (s *UploadService) GetUpload(ctx context.Context, uploadID entity.UploadID) (*entity.Upload, error) {
	upload, exists := s.uploads[uploadID]
	if !exists {
		return nil, ErrUploadNotFound
	}

	return upload, nil
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

func (s *UploadService) GetServiceConfiguration() Config {
	return s.config
}
