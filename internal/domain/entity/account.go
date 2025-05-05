package entity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"

	log "github.com/sirupsen/logrus"
)

type ServiceAccount struct {
	ID         string
	ProviderID string
	Region     string
	AccessKey  string
	Secret     string
}

type Provider struct {
	ID       string
	Name     string
	Endpoint string
}

func NewAccountID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		log.Error("failed to generate id")
	}
	return hex.EncodeToString(id)
}

func NewServiceAccount(id string, providerID string, region string, keyId string, secret string) *ServiceAccount {
	return &ServiceAccount{
		ID:         id,
		ProviderID: providerID,
		Region:     region,
		AccessKey:  keyId,
		Secret:     secret,
	}
}

func NewProvider(id string, name string, endpoint string) *Provider {
	return &Provider{
		ID:       id,
		Name:     name,
		Endpoint: endpoint,
	}
}

type AccountRepository interface {
	Add(ctx context.Context, account *ServiceAccount) error
	GetByID(ctx context.Context, accountID string) (*ServiceAccount, error)
	ListByProvider(ctx context.Context, provider string) ([]*ServiceAccount, error)
	List(ctx context.Context) ([]*ServiceAccount, error)
}

type ProviderRepository interface {
	GetByID(ctx context.Context, providerID string) (*Provider, error)
	List(ctx context.Context) ([]*Provider, error)
}
