package yandex

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BucketRepository struct {
	client *s3.Client
}

func NewBucketRepository(client *s3.Client) *BucketRepository {
	return &BucketRepository{
		client: client,
	}
}

func (r *BucketRepository) List(ctx context.Context) (*[]string, error) {
	result, err := r.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		fmt.Print("failed to list buckets")
		return nil, err
	}

	buckets := make([]string, len(result.Buckets))
	for index, bucket := range result.Buckets {
		buckets[index] = *bucket.Name
	}
	return &buckets, nil
}
