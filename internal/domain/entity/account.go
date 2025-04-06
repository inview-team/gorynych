package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"

	log "github.com/sirupsen/logrus"
)

type ServiceAccount struct {
	ID       string
	Provider Provider
	KeyID    string
	Secret   string
}

type Provider string

const (
	Yandex Provider = "yandex"
)

func NewAccountID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		log.Error("failed to generate id")
	}
	return hex.EncodeToString(id)
}

func NewServiceAccount(id string, provider Provider, keyId string, secret string) *ServiceAccount {
	return &ServiceAccount{
		ID:       id,
		Provider: provider,
		KeyID:    keyId,
		Secret:   secret,
	}
}

type AccountRepository interface {
	Add(ctx context.Context, account *ServiceAccount) error
	GetByID(ctx context.Context, accountID string) (*ServiceAccount, error)
	ListByProvider(ctx context.Context, provider Provider) ([]*ServiceAccount, error)
}
