package application

import (
	"github.com/inview-team/gorynych/internal/application/object"
	"github.com/inview-team/gorynych/internal/entities"
)

type Application struct {
	ObjectApplication object.Application
}

func New(ObjectRepository entities.ObjectRepository) *Application {
	return &Application{
		ObjectApplication: *object.New(ObjectRepository),
	}
}
