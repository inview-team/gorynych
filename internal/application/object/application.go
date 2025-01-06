package object

import "github.com/inview-team/gorynych/internal/entities"

type Application struct {
	oRepo entities.ObjectRepository
}

func New(oRepo entities.ObjectRepository) *Application {
	return &Application{
		oRepo: oRepo,
	}
}
