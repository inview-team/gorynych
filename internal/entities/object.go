package entities

import "context"

type Object struct {
	ID      ObjectID
	Name    string
	Size    string
	Hash    string
	Buckets []BucketID
}

type ObjectID string

type ObjectRepository interface {
	Get(ctx context.Context, id string) (*Object, error)
	Save(ctx context.Context, id string, payload []byte)
	List(ctx context.Context) ([]*Object, error)
	Delete(ctx context.Context, id string) error
}
