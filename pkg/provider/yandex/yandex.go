package yandex

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/inview-team/gorynych/internal/entities"
)

const Provider string = "yandex"

type YandexStorage struct {
	id       string
	s3Client *s3.Client
}

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
		s3Client: client,
	}, nil
}

func (s *YandexStorage) CreateUpload(ctx context.Context, bucketID entities.BucketID, object entities.Object) (string, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(string(bucketID)),
		Key:      aws.String(string(object.ID)),
		Metadata: object.Metadata,
	}

	resp, err := s.s3Client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create upload: %v", err)
	}

	uploadId := *resp.UploadId
	return uploadId, nil
}

func (s *YandexStorage) ListBuckets(ctx context.Context) ([]*entities.Bucket, error) {
	result, err := s.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Print("failed to list buckets")
		return nil, err
	}

	buckets := make([]string, len(result.Buckets))
	for index, bucket := range result.Buckets {
		buckets[index] = *bucket.Name
	}
	fmt.Println(buckets)
	return nil, nil
}
