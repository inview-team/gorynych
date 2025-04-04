package yandex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/domain/service"
)

const ProviderID string = "yandex"

type YandexStorage struct {
	mu       sync.Mutex
	s3Client *s3.Client
}

func New(ctx context.Context, id, secret string) (*YandexStorage, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(id, secret, "")),
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
		s3Client: client,
	}, nil
}

func (s *YandexStorage) GetProviderID(ctx context.Context) string {
	return ProviderID
}

func (s *YandexStorage) Create(ctx context.Context, storageID string, id string, metadata map[string]string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	input := &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(storageID),
		Key:      aws.String(id),
		Metadata: metadata,
	}

	resp, err := s.s3Client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	return *resp.UploadId, nil
}

func (s *YandexStorage) WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	input := &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(objectID),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(position)),
		Body:       bytes.NewReader(data),
	}

	resp, err := s.s3Client.UploadPart(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk: %v", err)
	}
	return *resp.ETag, nil
}

func (s *YandexStorage) FinishUpload(ctx context.Context, upload *entity.Upload) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := make([]types.CompletedPart, len(upload.Parts))
	for i, part := range upload.Parts {
		partNumber := part.Position
		parts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(partNumber)),
			ETag:       aws.String(part.ID),
		}
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(upload.Storage.Bucket),
		Key:             aws.String(upload.ObjectID),
		UploadId:        aws.String(upload.ID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: parts},
	}

	_, err := s.s3Client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to complete chunk: %v", err)
	}

	return nil
}

func (s *YandexStorage) ListBuckets(ctx context.Context) ([]string, error) {
	result, err := s.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Print("failed to list buckets")
		return nil, err
	}

	var buckets []string
	for _, bucket := range result.Buckets {
		buckets = append(buckets, *bucket.Name)
	}

	return buckets, nil
}

func (s *YandexStorage) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	_, err := s.s3Client.HeadBucket(ctx, input)
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
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
