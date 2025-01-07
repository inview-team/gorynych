package application

import (
	"github.com/inview-team/gorynych/internal/entities"
	"github.com/inview-team/gorynych/internal/service/object"
)

type Application struct {
	ObjectService object.Service
}

func New(ObjectRepository entities.ObjectRepository) *Application {
	return &Application{
		ObjectService: *object.New(ObjectRepository),
	}
}
