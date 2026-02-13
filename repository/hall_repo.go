package repository

import (
	"context"
	"cinema-system/model"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HallRepo struct {
	coll *mongo.Collection
}

func NewHallRepo(ctx context.Context, client *mongo.Client, dbName string) (*HallRepo, error) {
	coll := client.Database(dbName).Collection("halls")
	r := &HallRepo{coll: coll}
	if err := r.seedIfEmpty(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *HallRepo) seedIfEmpty(ctx context.Context) error {
	count, err := r.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seed := []model.Hall{
		{Name: "Зал 1", Capacity: 96, Rows: 8, SeatsPerRow: 12},
		{Name: "Зал 2", Capacity: 64, Rows: 8, SeatsPerRow: 8},
		{Name: "Зал 3", Capacity: 120, Rows: 10, SeatsPerRow: 12},
		{Name: "VIP Зал", Capacity: 24, Rows: 4, SeatsPerRow: 6},
	}
	for i := range seed {
		if _, err := r.Create(&seed[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *HallRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.Hall
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

func (r *HallRepo) Create(h *model.Hall) (*model.Hall, error) {
	ctx := context.Background()
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	h.ID = id
	if _, err := r.coll.InsertOne(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (r *HallRepo) GetByID(id int) (*model.Hall, error) {
	ctx := context.Background()
	var h model.Hall
	err := r.coll.FindOne(ctx, bson.D{{Key: "id", Value: id}}).Decode(&h)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *HallRepo) GetAll() ([]*model.Hall, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Hall
	for cur.Next(ctx) {
		var h model.Hall
		if err := cur.Decode(&h); err != nil {
			return nil, err
		}
		out = append(out, &h)
	}
	return out, cur.Err()
}
