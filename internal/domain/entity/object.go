package entity

import (
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
	Bucket   string
	Metadata map[string]string
}

type ObjectID string

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
