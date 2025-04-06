package entity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Replication struct {
	ID    ReplicationID
	Rules map[string]string
}

type ReplicationID string

func NewReplicationID() ReplicationID {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		fmt.Print("failed to generate id")
	}
	return ReplicationID(hex.EncodeToString(id))
}

type ReplicationRepository interface {
}
