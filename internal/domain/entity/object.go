package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Object struct {
	ID       ObjectID
	Name     string
	Size     int64
	Metadata map[string]string
}

type ObjectID string

func NewObjectID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return hex.EncodeToString(id)
}

type ObjectRepository interface {
	Create(ctx context.Context, bucket string, id string, metadata map[string]string) (string, error)
	WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data *[]byte) (string, error)
	FinishUpload(ctx context.Context, bucket, uploadID, objectID string, parts []UploadPart) error
	ListBuckets(ctx context.Context) ([]string, error)
	IsBucketExist(ctx context.Context, bucket string) (bool, error)
	GetObject(ctx context.Context, bucket string, objectID string) (*Object, error)
	DownloadObject(ctx context.Context, bucket string, objectID string, startOffset int64, endOffset int64) (*[]byte, error)
	StreamWritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data io.ReadCloser) (string, error)
	StreamDownloadObject(ctx context.Context, bucket string, objectID string, startOffset int64, endOffset int64) (io.ReadCloser, error)
}
