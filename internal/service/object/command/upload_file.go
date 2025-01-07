package command

import "github.com/inview-team/gorynych/internal/entities"

type UploadFileCommand struct {
	oRepo *entities.ObjectRepository
}

func NewUploadFileCommand(oRepo *entities.ObjectRepository) UploadFileCommand {
	return UploadFileCommand{
		oRepo: oRepo,
	}
}

func (c *UploadFileCommand) Execute(ctx, name string)
