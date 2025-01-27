package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Upload struct {
	ID        UploadID
	Size      int64
	Offset    int64
	Status    UploadStatus
	Metadata  map[string]string
	StorageID StorageID
}

type UploadID string

func NewUploadID() UploadID {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return UploadID(hex.EncodeToString(id))
}

type StorageID string

type UploadStatus int

const (
	Active UploadStatus = iota + 1
	Complete
	Expired
	Failed
)

func NewUpload(size int64, metadata map[string]string) *Upload {
	id := NewUploadID()
	return &Upload{
		ID:       id,
		Size:     size,
		Metadata: metadata,
		Offset:   0,
		Status:   Active,
	}
}

func (u *Upload) URL(baseURL string) string {
	return fmt.Sprintf("%s/%s", baseURL, u.ID)
}

func (u *Upload) SetOffset(offset int64) {
	u.Offset = offset
}

func (u *Upload) SetStorage(storageID StorageID) {
	u.StorageID = storageID
}

type UploadRepository interface {
	Create(ctx context.Context, upload *Upload) error
	WriteChunk(ctx context.Context, id UploadID, offset int64, data []byte) error
	Get(ctx context.Context, id string)
	FinishUpload(ctx context.Context, id UploadID) error
}
