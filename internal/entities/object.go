package entities

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

type Object struct {
	ID       ObjectID
	Name     string
	Size     int64
	Metadata map[string]string
	Buckets  []BucketID
}

type ObjectID string

type ObjectRepository interface {
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	CreateUpload(ctx context.Context, bucket string, info *UploadInfo) (UploadID, error)
	UploadPart(ctx context.Context, id UploadID, offset int64, src io.Reader) (int64, error)
}

type UploadInfo struct {
	ID             UploadID
	Size           int64
	Offset         int64
	Metadata       map[string]string
	IsPartial      bool
	IsFinal        bool
	PartialUploads []string
	BucketID       BucketID
}

func NewUploadInfo(size int64, metadata map[string]string, isPartial bool) *UploadInfo {
	info := UploadInfo{}
	info.Size = size
	info.Offset = 0
	info.IsPartial = isPartial
	if isPartial {
		info.IsFinal = false
	} else {
		info.IsFinal = true
	}
	info.PartialUploads = make([]string, 0)
	return &info
}

type UploadID string

func NewObjectID() ObjectID {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return ObjectID(hex.EncodeToString(id))
}

type Bucket struct {
	ID        BucketID
	Name      string
	StorageID string
}

type BucketID string

func NewBucketID(storageID string, bucketName string) BucketID {
	return BucketID(fmt.Sprintf("%s.%s", base64.StdEncoding.EncodeToString([]byte(storageID)), base64.StdEncoding.EncodeToString([]byte(bucketName))))
}
