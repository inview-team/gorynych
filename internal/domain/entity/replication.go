package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Replication struct {
	ID    ReplicationID
	Rules map[string]string
}

type ReplicationID string

func NewReplicationID() ReplicationID {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return ReplicationID(hex.EncodeToString(id))
}

type ReplicationRepository interface {
	GetObject(ctx context.Context, id ObjectID, bucketName string) (*Object, error)
	CopyObject(ctx context.Context, sourceObject ObjectID, targetObject ObjectID, sourceBucket string, targetBucket string) error
	ListBuckets(ctx context.Context) ([]*Bucket, error)
}
