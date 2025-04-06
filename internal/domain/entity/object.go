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
	Bucket   string
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
	WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data []byte) (string, error)
	FinishUpload(ctx context.Context, upload *Upload) error
	ListBuckets(ctx context.Context) ([]string, error)
	IsBucketExist(ctx context.Context, bucket string) (bool, error)
	GetProviderID(ctx context.Context) string
}
