package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

type Task struct {
	ID     string
	Start  time.Time
	End    time.Time
	Type   TaskType
	Status TaskStatus
}

type TaskStatus int

const (
	TaskCreated TaskStatus = iota + 1
	TaskCompleted
	TaskFailed
)

type TaskType int

const (
	Replication TaskType = iota + 1
)

func NewTaskID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return hex.EncodeToString(id)
}

type TaskRepository interface {
	Add(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, taskID string) (*Task, error)
	Update(ctx context.Context, task *Task) error
}
