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

type YandexStorage struct {
	Object *ObjectRepository
	Bucket *BucketRepository
}

type Credentials struct {
	AccessKeyID     string `yaml:"aws_access_key_id"`
	SecretAccessKey string `yaml:"aws_secret_access_key "`
}

func New(ctx context.Context, creds Credentials) (*YandexStorage, error) {
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
		Object: NewObjectRepository(client),
		Bucket: NewBucketRepository(client),
	}, nil
}

// Save implements entities.ObjectRepository.
func (y *YandexStorage) Save(ctx context.Context, id string, name string, contentType string, payload []byte, bucket string) {
	panic("unimplemented")
}

// Delete implements entities.ObjectRepository.
func (y *YandexStorage) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// Get implements entities.ObjectRepository.
func (y *YandexStorage) Get(ctx context.Context, id string) (*entities.Object, error) {
	panic("unimplemented")
}

// List implements entities.ObjectRepository.
func (y *YandexStorage) List(ctx context.Context) ([]*entities.Object, error) {
	panic("unimplemented")
}
