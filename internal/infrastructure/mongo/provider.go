package mongo

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProviderRepository struct {
	coll *mongo.Collection
}

func NewProviderRepository(ctx context.Context, client *Client) (*ProviderRepository, error) {
	coll := client.Database.Collection("providers")

	for _, provider := range model.Providers {
		err := coll.FindOne(ctx, bson.M{"_id": provider.ID}).Decode(&model.Providers)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, err
			}
		}
		_, err = coll.InsertOne(ctx, provider)
		if err != nil {
			return nil, err
		}
	}

	return &ProviderRepository{
		coll: coll,
	}, nil
}

func (r *ProviderRepository) GetByID(ctx context.Context, providerID string) (*entity.Provider, error) {

	result := r.coll.FindOne(ctx, bson.M{"_id": providerID})

	var mProvider model.Provider
	err := result.Decode(&mProvider)
	if err != nil {
		return nil, err
	}

	return mProvider.ToEntity(), nil
}

func (r *ProviderRepository) List(ctx context.Context) ([]*entity.Provider, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var providers []*entity.Provider
	for cursor.Next(ctx) {
		var mProvider model.Provider
		if err := cursor.Decode(&mProvider); err != nil {
			return nil, err
		}
		providers = append(providers, mProvider.ToEntity())
	}
	return providers, nil
}
