package model

import "github.com/inview-team/gorynych/internal/domain/entity"

type Account struct {
	ID       string `bson:"_id"`
	Provider string `bson:"provider"`
	KeyID    string `bson:"key_id"`
	Secret   string `bson:"secret"`
}

func NewAccount(account *entity.ServiceAccount) *Account {
	return &Account{
		ID:       account.ID,
		Provider: string(account.Provider),
		KeyID:    account.KeyID,
		Secret:   account.Secret,
	}
}

func (m *Account) ToEntity() *entity.ServiceAccount {
	return &entity.ServiceAccount{
		ID:       m.ID,
		Provider: entity.Provider(m.Provider),
		KeyID:    m.KeyID,
		Secret:   m.Secret,
	}
}
