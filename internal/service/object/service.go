package object

import (
	"context"
	"fmt"

	"github.com/inview-team/gorynych/internal/entities"
)

type Commands struct {
}

type Service struct {
	storages map[string]entities.ObjectRepository
	uploads  map[string]*entities.UploadInfo
}

func New(storages []*entities.ObjectRepository) *Service {
	return &Service{
		storages: make(map[string]entities.ObjectRepository),
		uploads:  make(map[string]*entities.UploadInfo),
	}
}

func (s *Service) RegisterStorage(ctx context.Context, id string, storage entities.ObjectRepository) error {
	if _, ok := s.storages[id]; ok {
		return ErrRepositoryExists
	}
	s.storages[id] = storage
	return nil
}

func (s *Service) DeregisterStorage(ctx context.Context, id string) error {
	if _, ok := s.storages[id]; !ok {
		return ErrRepositoryNotFound
	}
	delete(s.storages, id)
	return nil
}

func (s *Service) CreateUpload(ctx context.Context, size int64, metadata map[string]string, isPartial bool) (entities.UploadID, error) {
	uploadInfo := entities.NewUploadInfo(size, metadata, isPartial)
	bucket, err := s.chooseUploadBucket(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create upload")
	}
	uploadID, err := s.storages[bucket.StorageID].CreateUpload(ctx, bucket.Name, uploadInfo)
	if err != nil {
		return "", fmt.Errorf("failed to create upload")
	}
	return uploadID, nil
}

func (s *Service) chooseUploadBucket(ctx context.Context) (*entities.Bucket, error) {
	for _, client := range s.storages {
		cBuckets, err := client.ListBuckets(ctx)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to list buckets: %v", err.Error()))
			continue
		}
		return cBuckets[0], nil
	}
	return nil, ErrNoAvailableBuckets
}
