package yandex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/inview-team/gorynych/internal/entities"
)

type ObjectRepository struct {
	client *s3.Client
}

func NewObjectRepository(client *s3.Client) *ObjectRepository {
	return &ObjectRepository{
		client: client,
	}
}

// Delete implements entities.ObjectRepository.
func (y *ObjectRepository) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// Get implements entities.ObjectRepository.
func (y *ObjectRepository) Get(ctx context.Context, id string) (*entities.Object, error) {
	panic("unimplemented")
}

// List implements entities.ObjectRepository.
func (y *ObjectRepository) List(ctx context.Context) ([]*entities.Object, error) {
	panic("unimplemented")
}

// Save implements entities.ObjectRepository.
func (y *ObjectRepository) Save(ctx context.Context, id string, payload []byte) {
	panic("unimplemented")
}
