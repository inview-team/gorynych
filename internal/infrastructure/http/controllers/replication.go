package controllers

import "github.com/inview-team/gorynych/internal/domain/entity"

type File struct {
	ID string `json:"id"`
}

func (r *File) ToEntity() entity.ObjectID {
	return entity.ObjectID(r.ID)
}
