package yandex

import (
	"context"

	"github.com/inview-team/gorynych/internal/entities"
)

type YandexStorage struct {
}

func New() *YandexStorage {
	return &YandexStorage{}
}

// Delete implements entities.ObjectRepository.
func (y *YandexStorage) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// Get implements entities.ObjectRepository.
func (y *YandexStorage) Get(ctx context.Context, id string) (*entities.Object, error) {
	panic("unimplemented")
}

// List implements entities.ObjectRepository.
func (y *YandexStorage) List(ctx context.Context) ([]*entities.Object, error) {
	panic("unimplemented")
}

// Save implements entities.ObjectRepository.
func (y *YandexStorage) Save(ctx context.Context, id string, payload []byte) {
	panic("unimplemented")
}
