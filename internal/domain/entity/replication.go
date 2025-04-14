package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type ReplicationTask struct {
	ID                  string
	ObjectID            string
	Priority            Priority
	UploadInformation   *Upload
	DownloadInformation *Download
}

type Priority int

const (
	ManualReplicate Priority = iota
	LostObject
)

func NewReplicationTask(id string, objectID string, priority Priority, downloadInfo *Download, uploadInfo *Upload) *ReplicationTask {
	return &ReplicationTask{
		ID:                  id,
		ObjectID:            objectID,
		Priority:            priority,
		DownloadInformation: downloadInfo,
		UploadInformation:   uploadInfo,
	}
}

type Download struct {
	Storage Storage
}

func NewDownload(storage Storage) *Download {
	return &Download{
		Storage: storage,
	}
}

func NewReplicationID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return hex.EncodeToString(id)
}

type ReplicationRepository interface {
	Add(ctx context.Context, task *ReplicationTask) error
	GetByID(ctx context.Context, taskID string) (*ReplicationTask, error)
	Update(ctx context.Context, task *ReplicationTask) error
	ListByPriority(ctx context.Context, priority Priority) ([]string, error)
}
