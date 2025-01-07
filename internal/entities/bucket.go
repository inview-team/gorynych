package entities

import "context"

type Bucket struct {
	ID       BucketID
	Name     string
	Provider string
}

type BucketID string

type Status int

const (
	Available Status = iota + 1
	UnAvailable
)

type BucketRepository interface {
	List(ctx context.Context) ([]*Bucket, error)
}
