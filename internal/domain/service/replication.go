package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/inview-team/gorynych/internal/domain/entity"
)

type ReplicationService struct {
	mu       sync.Mutex
	storages map[string]entity.ReplicationRepository
}

func NewReplicationService() *ReplicationService {
	return &ReplicationService{
		storages: make(map[string]entity.ReplicationRepository),
	}
}

func (s *ReplicationService) RegisterStorage(ctx context.Context, id string, storage entity.ReplicationRepository) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storages[id]; ok {
		return ErrRepositoryExists
	}
	s.storages[id] = storage
	return nil
}

func (s *ReplicationService) DeregisterStorage(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storages[id]; !ok {
		return ErrRepositoryNotFound
	}
	delete(s.storages, id)
	return nil
}

func (s *ReplicationService) CreateReplication(ctx context.Context, objectID entity.ObjectID) (entity.ObjectID, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	storageID, object, err := s.findObject(ctx, objectID)
	if err != nil {
		return "", "", ErrObjectNotFound
	}
	fmt.Printf("Object: %v", object)
	targetBucket, err := s.chooseTargetBucket(ctx, storageID, object.Bucket)
	if err != nil {
		return "", "", ErrNoAvailableBuckets
	}

	targetObjectID := entity.NewObjectID()
	err = s.storages[storageID].CopyObject(ctx, object.ID, targetObjectID, object.Bucket, targetBucket.Name)
	if err != nil {
		return "", "", fmt.Errorf("failed to create replication: %v", err)
	}
	return targetObjectID, targetBucket.Name, nil
}

func (s *ReplicationService) findObject(ctx context.Context, objectID entity.ObjectID) (string, *entity.Object, error) {
	if len(s.storages) == 0 {
		return "", nil, errors.New("no available repositories")
	}

	for id, storage := range s.storages {
		buckets, err := storage.ListBuckets(ctx)
		if err != nil {
			continue
		}
		fmt.Println(buckets)
		for _, bucket := range buckets {
			fmt.Printf("Check bucket %s\n", bucket.Name)
			object, err := storage.GetObject(ctx, objectID, bucket.Name)
			if errors.Is(err, ErrObjectNotFound) {
				continue
			}

			return id, object, nil
		}
	}
	return "", nil, ErrObjectNotFound
}

func (s *ReplicationService) chooseTargetBucket(ctx context.Context, id string, sourceBucket string) (*entity.Bucket, error) {
	buckets, err := s.storages[id].ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to choose bucket: failed to list buckets")
	}

	for _, bucket := range buckets {
		if bucket.Name == sourceBucket {
			continue
		}
		return bucket, nil
	}
	return nil, fmt.Errorf("failed to choose bucket: failed to find bucket")
}
