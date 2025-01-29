package views

import "github.com/inview-team/gorynych/internal/domain/entity"

type ReplicatedFile struct {
	ID     string `json:"id"`
	Bucket string `json:"bucket"`
}

func NewReplicatedFile(id entity.ObjectID, bucket string) *ReplicatedFile {
	return &ReplicatedFile{
		ID:     string(id),
		Bucket: bucket,
	}
}
