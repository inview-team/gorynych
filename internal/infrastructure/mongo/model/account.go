package model

import "github.com/inview-team/gorynych/internal/domain/entity"

type Account struct {
	ID         string `bson:"_id"`
	ProviderID string `bson:"provider_id"`
	Region     string `bson:"region"`
	AccessKey  string `bson:"access_key"`
	Secret     string `bson:"secret"`
}

func NewAccount(account *entity.ServiceAccount) *Account {
	return &Account{
		ID:         account.ID,
		ProviderID: account.ProviderID,
		Region:     account.Region,
		AccessKey:  account.AccessKey,
		Secret:     account.Secret,
	}
}

func (m *Account) ToEntity() *entity.ServiceAccount {
	return &entity.ServiceAccount{
		ID:         m.ID,
		ProviderID: m.ProviderID,
		Region:     m.Region,
		AccessKey:  m.AccessKey,
		Secret:     m.Secret,
	}
}
