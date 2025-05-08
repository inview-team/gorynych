package entity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

type ReplicationTask struct {
	ID            string
	ObjectID      string
	SourceStorage Storage
	TargetStorage Storage
}

type ReplicationResult struct {
	ID    string
	Start time.Time
	End   time.Time
	Error error
}

func NewReplicationTask(objectID string, sStorage Storage, tStorage Storage) *ReplicationTask {
	return &ReplicationTask{
		ObjectID:      objectID,
		SourceStorage: sStorage,
		TargetStorage: tStorage,
	}
}

func NewReplicationID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return hex.EncodeToString(id)
}
