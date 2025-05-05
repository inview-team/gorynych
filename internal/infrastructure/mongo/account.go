package mongo

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AccountRepository struct {
	coll *mongo.Collection
}

func NewAccountRepository(client *Client) *AccountRepository {
	return &AccountRepository{
		coll: client.Database.Collection("accounts"),
	}
}

func (r *AccountRepository) Add(ctx context.Context, account *entity.ServiceAccount) error {
	mAccount := model.NewAccount(account)
	_, err := r.coll.InsertOne(ctx, mAccount)
	if err != nil {
		return err
	}
	return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, accountID string) (*entity.ServiceAccount, error) {

	result := r.coll.FindOne(ctx, bson.M{"_id": accountID})

	var mAccount model.Account
	err := result.Decode(&mAccount)
	if err != nil {
		return nil, err
	}

	return mAccount.ToEntity(), nil
}

func (r *AccountRepository) ListByProvider(ctx context.Context, provider string) ([]*entity.ServiceAccount, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"provider_id": provider})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []*entity.ServiceAccount
	for cursor.Next(ctx) {
		var mAccount model.Account
		if err := cursor.Decode(&mAccount); err != nil {
			return nil, err
		}
		accounts = append(accounts, mAccount.ToEntity())
	}
	return accounts, nil
}

func (r *AccountRepository) List(ctx context.Context) ([]*entity.ServiceAccount, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []*entity.ServiceAccount
	for cursor.Next(ctx) {
		var mAccount model.Account
		if err := cursor.Decode(&mAccount); err != nil {
			return nil, err
		}
		accounts = append(accounts, mAccount.ToEntity())
	}
	return accounts, nil
}
