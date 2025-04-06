package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Upload struct {
	ID       string
	ObjectID string
	Size     int64
	Offset   int64
	Storage  Storage
	Parts    []UploadPart
	Status   UploadStatus
}

type UploadPart struct {
	ID       string
	Position int
}

type Storage struct {
	ProviderID string
	Bucket     string
}

func NewStorageID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return hex.EncodeToString(id)
}

type UploadStatus int

const (
	Active UploadStatus = iota + 1
	Complete
	Expired
	Failed
)

func NewUpload(id string, objectID string, size int64, offset int64, status UploadStatus, parts []UploadPart, storage Storage) *Upload {
	return &Upload{
		ID:       id,
		ObjectID: objectID,
		Size:     size,
		Offset:   offset,
		Parts:    parts,
		Storage:  storage,
		Status:   status,
	}
}

func (u *Upload) SetOffset(offset int64) {
	u.Offset = offset
}

func (u *Upload) AddPartial(partialID string, position int) {
	u.Parts = append(u.Parts, UploadPart{ID: partialID, Position: position})
}

type UploadRepository interface {
	Add(ctx context.Context, upload *Upload) error
	GetByID(ctx context.Context, uploadID string) (*Upload, error)
	Update(ctx context.Context, upload *Upload) error
}

type StorageRepository interface {
	Create(ctx context.Context, bucket string, id string, metadata map[string]string) (string, error)
	WritePart(ctx context.Context, bucket string, uploadID string, objectID string, position int, data []byte) (string, error)
	FinishUpload(ctx context.Context, upload *Upload) error
	ListBuckets(ctx context.Context) ([]string, error)
	IsBucketExist(ctx context.Context, bucket string) (bool, error)
	GetProviderID(ctx context.Context) string
}
