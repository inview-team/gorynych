package model

import "github.com/inview-team/gorynych/internal/domain/entity"

type Provider struct {
	ID       string `bson:"_id"`
	Name     string `bson:"name"`
	Endpoint string `bson:"endpoint"`
}

var Providers = []Provider{
	{
		ID:       "1",
		Name:     "Yandex Cloud",
		Endpoint: "https://storage.yandexcloud.net",
	},
	{
		ID:       "2",
		Name:     "Timeweb",
		Endpoint: "https://s3.twcstorage.ru",
	},
}

func (m *Provider) ToEntity() *entity.Provider {
	return &entity.Provider{
		ID:       m.ID,
		Name:     m.Name,
		Endpoint: m.Endpoint,
	}
}
