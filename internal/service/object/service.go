package object

import "github.com/inview-team/gorynych/internal/entities"

type Commands struct {
}

type Service struct {
	oRepo entities.ObjectRepository
}

func New(oRepo entities.ObjectRepository) *Service {
	return &Service{
		oRepo: oRepo,
	}
}
