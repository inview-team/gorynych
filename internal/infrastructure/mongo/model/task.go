package model

import (
	"time"

	"github.com/inview-team/gorynych/internal/domain/entity"
)

const (
	layout = "2006-01-02 15:04:05"
)

type Task struct {
	ID     string `bson:"_id"`
	Start  string `bson:"start"`
	End    string `bson:"end"`
	Type   int    `bson:"type"`
	Status int    `bson:"status"`
}

func NewTask(task *entity.Task) *Task {
	return &Task{
		ID:     task.ID,
		Start:  task.Start.String(),
		End:    task.End.String(),
		Type:   int(task.Type),
		Status: int(task.Status),
	}
}

func (m *Task) ToEntity() *entity.Task {
	start, _ := time.Parse(layout, m.Start)
	end, _ := time.Parse(layout, m.End)
	return &entity.Task{
		ID:     m.ID,
		Start:  start,
		End:    end,
		Type:   entity.TaskType(m.Type),
		Status: entity.TaskStatus(m.Status),
	}
}
