package repository

import (
	"context"
	"cinema-system/model"

	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepo — репозиторий пользователей на MongoDB.
type UserRepo struct {
	coll *mongo.Collection
}

func NewUserRepo(ctx context.Context, client *mongo.Client, dbName string) *UserRepo {
	coll := client.Database(dbName).Collection("users")
	return &UserRepo{coll: coll}
}

// nextID ищет максимальный id и возвращает следующий.
func (r *UserRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.User
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

// Create сохраняет нового пользователя.
func (r *UserRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	u.ID = id
	if _, err := r.coll.InsertOne(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetByEmail возвращает пользователя по email или nil, если не найден.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.coll.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CountByRole считает пользователей с указанной ролью.
func (r *UserRepo) CountByRole(ctx context.Context, role string) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.D{{Key: "role", Value: role}})
}

