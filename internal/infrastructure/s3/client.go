package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/inview-team/gorynych/internal/domain/entity"

	log "github.com/sirupsen/logrus"
)

type ClientS3 struct {
	s3Client *s3.Client
}

func New(ctx context.Context, endpoint, region, accessKey, secret string) (*ClientS3, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(accessKey, secret, "")),
	)

	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Region = region
	})

	return &ClientS3{
		s3Client: client,
	}, nil
}

func (s *ClientS3) Create(ctx context.Context, storageID string, id string, metadata map[string]string) (string, error) {
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

func (s *ClientS3) WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data *[]byte) (string, error) {
	input := &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(objectID),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(position)),
		Body:       bytes.NewReader(*data),
	}

	resp, err := s.s3Client.UploadPart(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk: %v", err)
	}

	return *resp.ETag, nil
}

func (s *ClientS3) FinishUpload(ctx context.Context, bucket, uploadID, objectID string, uploadedParts []entity.UploadPart) error {
	parts := make([]types.CompletedPart, len(uploadedParts))
	for i, part := range uploadedParts {
		partNumber := part.Position
		parts[i] = types.CompletedPart{
			PartNumber: aws.Int32(int32(partNumber)),
			ETag:       aws.String(part.ID),
		}
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(objectID),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{Parts: parts},
	}

	_, err := s.s3Client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to complete chunk: %v", err)
	}

	return nil
}

func (s *ClientS3) ListBuckets(ctx context.Context) ([]string, error) {
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

func (s *ClientS3) IsBucketExist(ctx context.Context, bucket string) (bool, error) {
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

// DownloadObject implements entity.ObjectRepository.
func (s *ClientS3) DownloadObject(ctx context.Context, bucket string, objectID string, startOffset int64, endOffset int64) (*[]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectID),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", startOffset, endOffset)),
	}

	output, err := s.s3Client.GetObject(ctx, input)
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	data, _ := io.ReadAll(output.Body)
	defer output.Body.Close()
	return &data, nil
}

// GetObject implements entity.ObjectRepository.
func (s *ClientS3) GetObject(ctx context.Context, bucket string, objectID string) (*entity.Object, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectID),
	}

	output, err := s.s3Client.HeadObject(ctx, input)
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			log.Errorf(err.Error())
			return nil, nil
		}
		return nil, err
	}

	return &entity.Object{
		ID:       entity.ObjectID(objectID),
		Name:     objectID,
		Size:     *output.ContentLength,
		Metadata: output.Metadata,
	}, nil
}

func (s *ClientS3) StreamDownloadObject(ctx context.Context, bucket string, objectID string, startOffset int64, endOffset int64) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectID),
		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", startOffset, endOffset)),
	}

	output, err := s.s3Client.GetObject(ctx, input)
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	return output.Body, nil
}

func (s *ClientS3) StreamWritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, reader io.ReadCloser) (string, error) {
	input := &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(objectID),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(position)),
		Body:       reader,
	}

	resp, err := s.s3Client.UploadPart(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk: %v", err)
	}

	return *resp.ETag, nil
}
