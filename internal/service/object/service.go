package object

import (
	"github.com/inview-team/gorynych/internal/entities"
)

type Commands struct {
}

type Service struct {
	storages map[string]*entities.ObjectRepository
}

func New(storages []*entities.ObjectRepository) *Service {
	return &Service{
		storages: make(map[string]*entities.ObjectRepository),
	}
}

func (s *Service) RegisterStorage(id string, storage *entities.ObjectRepository) error {
	if _, ok := s.storages[id]; ok {
		return ErrRepositoryExists
	}
	s.storages[id] = storage
	return nil
}

func (s *Service) DeregisterStorage(id string) error {
	if _, ok := s.storages[id]; !ok {
		return ErrRepositoryNotFound
	}
	delete(s.storages, id)
	return nil
}
