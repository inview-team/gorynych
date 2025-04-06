package mongo

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UploadRepository struct {
	coll *mongo.Collection
}

func NewUploadRepository(client *Client) *UploadRepository {
	return &UploadRepository{
		coll: client.Database.Collection("uploads"),
	}
}

func (r *UploadRepository) Add(ctx context.Context, upload *entity.Upload) error {
	mUpload := model.NewUpload(upload)
	_, err := r.coll.InsertOne(ctx, mUpload)
	if err != nil {
		return err
	}
	return nil
}

func (r *UploadRepository) GetByID(ctx context.Context, objectId string) (*entity.Upload, error) {

	result := r.coll.FindOne(ctx, bson.M{"object_id": objectId})

	var mUpload model.Upload
	err := result.Decode(&mUpload)
	if err != nil {
		return nil, err
	}

	return mUpload.ToEntity(), nil
}

func (r *UploadRepository) Update(ctx context.Context, upload *entity.Upload) error {
	mUpload := model.NewUpload(upload)
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{
			"_id": bson.M{"$eq": upload.ID},
		},
		bson.M{"$set": mUpload},
	)

	if err != nil {
		return err
	}
	return nil
}
