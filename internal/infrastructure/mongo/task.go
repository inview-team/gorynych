package mongo

import (
	"context"

	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/infrastructure/mongo/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type TaskRepository struct {
	coll *mongo.Collection
}

func NewTaskRepository(client *Client) *TaskRepository {
	return &TaskRepository{
		coll: client.Database.Collection("tasks"),
	}
}

func (r *TaskRepository) Add(ctx context.Context, Task *entity.Task) error {
	mTask := model.NewTask(Task)
	_, err := r.coll.InsertOne(ctx, mTask)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID string) (*entity.Task, error) {

	result := r.coll.FindOne(ctx, bson.M{"_id": taskID})

	var mTask model.Task
	err := result.Decode(&mTask)
	if err != nil {
		return nil, err
	}

	return mTask.ToEntity(), nil
}

func (r *TaskRepository) Update(ctx context.Context, Task *entity.Task) error {
	mTask := model.NewTask(Task)
	_, err := r.coll.UpdateOne(
		ctx,
		bson.M{
			"_id": bson.M{"$eq": Task.ID},
		},
		bson.M{"$set": mTask},
	)

	if err != nil {
		return err
	}
	return nil
}
