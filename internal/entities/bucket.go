package entities

type Bucket struct {
	ID       BucketID
	Name     string
	Status   Status
	Provider Provider
}

type BucketID string

type Provider int

const (
	Yandex Provider = iota + 1
)

type Status int

const (
	Available Status = iota + 1
	UnAvailable
)
