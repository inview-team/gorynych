package yandex

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/pkg/storage/s3/yandex/model"
)

const Provider string = "yandex"

type YandexStorage struct {
	mu       sync.Mutex
	id       string
	s3Client *s3.Client
	uploads  map[entity.UploadID]*model.MultipartUpload
}

func (s *YandexStorage) GetID() string {
	return s.id
}

// Get implements entity.UploadRepository.
func (s *YandexStorage) Get(ctx context.Context, id string) {
	panic("unimplemented")
}

// WriteChunk implements entity.UploadRepository.

type S3Credentials struct {
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

func New(ctx context.Context, id string, creds S3Credentials) (*YandexStorage, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, "")),
	)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://storage.yandexcloud.net")
		o.Region = "ru-central1"
	})

	return &YandexStorage{
		id:       id,
		s3Client: client,
		uploads:  make(map[entity.UploadID]*model.MultipartUpload),
	}, nil
}

func (s *YandexStorage) Create(ctx context.Context, upload *entity.Upload) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	buckets, err := s.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload: %v", err)
	}

	index := 0

	input := &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(string(buckets[index].Name)),
		Key:      aws.String(string(upload.ID)),
		Metadata: upload.Metadata,
	}

	resp, err := s.s3Client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create upload: %v", err)
	}

	mpUpload := model.NewMultiPartUpload(*resp.UploadId, buckets[index].Name)
	s.uploads[upload.ID] = mpUpload
	fmt.Println(mpUpload)
	return nil
}

func (s *YandexStorage) WriteChunk(ctx context.Context, id entity.UploadID, offset int64, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	upload, exists := s.uploads[id]
	if !exists {
		return fmt.Errorf("failed to upload chunk: upload not found")
	}

	input := &s3.UploadPartInput{
		Bucket:     aws.String(upload.Bucket),
		Key:        aws.String(string(id)),
		UploadId:   aws.String(upload.ID),
		PartNumber: aws.Int32(int32(len(upload.Partials)) + 1),
		Body:       bytes.NewReader(data),
	}

	resp, err := s.s3Client.UploadPart(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload chunk: %v", err)
	}

	s.uploads[id].AddPartial(*resp.ETag)
	return nil
}

func (s *YandexStorage) FinishUpload(ctx context.Context, id entity.UploadID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	upload, exists := s.uploads[id]
	if !exists {
		return fmt.Errorf("failed to upload chunk: upload not found")
	}

	parts := make([]types.CompletedPart, len(upload.Partials))
	for i, tag := range upload.Partials {
		partNumber := i + 1
		parts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(partNumber)),
			ETag:       aws.String(tag),
		}
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(upload.Bucket),
		Key:             aws.String(string(id)),
		UploadId:        aws.String(upload.ID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: parts},
	}

	_, err := s.s3Client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to complete chunk: %v", err)
	}

	return nil
}

func (s *YandexStorage) ListBuckets(ctx context.Context) ([]*entity.Bucket, error) {
	result, err := s.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Print("failed to list buckets")
		return nil, err
	}

	buckets := make([]*entity.Bucket, len(result.Buckets))
	for index, bucket := range result.Buckets {
		buckets[index] = &entity.Bucket{
			ID:        entity.NewBucketID(s.id, *bucket.Name),
			Name:      *bucket.Name,
			StorageID: s.id,
		}
	}

	return buckets, nil
}

// GetObject implements entity.ReplicationRepository.
func (s *YandexStorage) GetObject(ctx context.Context, id entity.ObjectID, bucketName string) (*entity.Object, error) {
	input := &s3.GetObjectInput{
		Key:    aws.String(string(id)),
		Bucket: aws.String(bucketName),
	}
	resp, _ := s.s3Client.GetObject(ctx, input)
	if resp == nil {
		return nil, service.ErrObjectNotFound
	}

	return &entity.Object{
		ID:       id,
		Name:     resp.Metadata["name"],
		Size:     *resp.ContentLength,
		Metadata: resp.Metadata,
		Bucket:   bucketName,
	}, nil
}

// CopyObject implements entity.ReplicationRepository.
func (s *YandexStorage) CopyObject(ctx context.Context, sourceObject entity.ObjectID, targetObject entity.ObjectID, sourceBucket string, targetBucket string) error {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(targetBucket),
		Key:        aws.String(string(targetObject)),
		CopySource: aws.String(sourceBucket + "/" + string(sourceObject)),
	}

	_, err := s.s3Client.CopyObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to copy object: %v", err)
	}
	return nil
}
