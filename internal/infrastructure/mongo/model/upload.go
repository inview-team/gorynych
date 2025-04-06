package model

import (
	"github.com/inview-team/gorynych/internal/domain/entity"
)

type Upload struct {
	ID       string       `bson:"_id"`
	ObjectID string       `bson:"object_id"`
	Size     int64        `bson:"size"`
	Offset   int64        `bson:"offset"`
	Storage  Storage      `bson:"storage"`
	Parts    []UploadPart `bson:"parts"`
	Status   int          `bson:"status"`
}

type Storage struct {
	ProviderID string `bson:"provider_id"`
	Bucket     string `bson:"bucket"`
}

type UploadPart struct {
	ID       string `bson:"part_id"`
	Position int    `bson:"position"`
}

func NewUpload(upload *entity.Upload) *Upload {
	var parts []UploadPart
	for _, part := range upload.Parts {
		parts = append(parts, UploadPart{ID: part.ID, Position: part.Position})
	}

	return &Upload{
		ID:       upload.ID,
		ObjectID: upload.ObjectID,
		Size:     upload.Size,
		Offset:   upload.Offset,
		Storage: Storage{
			ProviderID: upload.Storage.ProviderID,
			Bucket:     upload.Storage.Bucket,
		},
		Parts:  parts,
		Status: int(upload.Status),
	}
}

func (m *Upload) ToEntity() *entity.Upload {
	var parts []entity.UploadPart
	for _, part := range m.Parts {
		parts = append(parts, entity.UploadPart{ID: part.ID, Position: part.Position})
	}

	status := entity.UploadStatus(m.Status)
	return &entity.Upload{
		ID:       m.ID,
		ObjectID: m.ObjectID,
		Size:     m.Size,
		Offset:   m.Offset,
		Storage: entity.Storage{
			ProviderID: m.Storage.ProviderID,
			Bucket:     m.Storage.Bucket,
		},
		Parts:  parts,
		Status: status,
	}
}
