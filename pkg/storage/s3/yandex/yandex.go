package yandex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inview-team/gorynych/internal/domain/entity"
)

const ProviderID string = "yandex"

type ClientYandex struct {
	s3Client *s3.Client
}

func New(ctx context.Context, id, secret string) (*ClientYandex, error) {
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

	return &ClientYandex{
		s3Client: client,
	}, nil
}

func (s *ClientYandex) GetProviderID(ctx context.Context) string {
	return ProviderID
}

func (s *ClientYandex) Create(ctx context.Context, storageID string, id string, metadata map[string]string) (string, error) {
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

func (s *ClientYandex) WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data []byte) (string, error) {
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

func (s *ClientYandex) FinishUpload(ctx context.Context, upload *entity.Upload) error {
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

func (s *ClientYandex) ListBuckets(ctx context.Context) ([]string, error) {
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

func (s *ClientYandex) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
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
