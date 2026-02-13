package repository

import (
	"context"
	"cinema-system/model"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionRepo struct {
	coll *mongo.Collection
}

func NewSessionRepo(ctx context.Context, client *mongo.Client, dbName string) (*SessionRepo, error) {
	coll := client.Database(dbName).Collection("sessions")
	r := &SessionRepo{coll: coll}
	if err := r.seedIfEmpty(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SessionRepo) seedIfEmpty(ctx context.Context) error {
	count, err := r.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// Несколько сеансов: фильм 1 в залах 1,2 в разное время
	seed := []model.Session{
		{MovieID: 1, HallID: 1, StartTime: "2026-02-15T11:00:00Z", Price: 2500},
		{MovieID: 1, HallID: 2, StartTime: "2026-02-15T14:30:00Z", Price: 2500},
		{MovieID: 2, HallID: 1, StartTime: "2026-02-15T16:00:00Z", Price: 2200},
		{MovieID: 2, HallID: 3, StartTime: "2026-02-15T19:00:00Z", Price: 2200},
		{MovieID: 3, HallID: 2, StartTime: "2026-02-15T12:00:00Z", Price: 2500},
		{MovieID: 4, HallID: 1, StartTime: "2026-02-15T21:00:00Z", Price: 1900},
	}
	for i := range seed {
		if _, err := r.Create(&seed[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *SessionRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.Session
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

func (r *SessionRepo) Create(s *model.Session) (*model.Session, error) {
	ctx := context.Background()
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	s.ID = id
	if _, err := r.coll.InsertOne(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SessionRepo) GetByID(id int) (*model.Session, error) {
	ctx := context.Background()
	var s model.Session
	err := r.coll.FindOne(ctx, bson.D{{Key: "id", Value: id}}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepo) GetAll() ([]*model.Session, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Session
	for cur.Next(ctx) {
		var s model.Session
		if err := cur.Decode(&s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, cur.Err()
}

func (r *SessionRepo) GetByMovieID(movieID int) ([]*model.Session, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{{Key: "movie_id", Value: movieID}},
		options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Session
	for cur.Next(ctx) {
		var s model.Session
		if err := cur.Decode(&s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, cur.Err()
}

// Update обновляет сеанс по ID.
func (r *SessionRepo) Update(s *model.Session) error {
	ctx := context.Background()
	_, err := r.coll.UpdateOne(ctx, bson.D{{Key: "id", Value: s.ID}}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "movie_id", Value: s.MovieID},
			{Key: "hall_id", Value: s.HallID},
			{Key: "start_time", Value: s.StartTime},
			{Key: "price", Value: s.Price},
		}},
	})
	return err
}

func (r *SessionRepo) GetByHallID(hallID int) ([]*model.Session, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{{Key: "hall_id", Value: hallID}},
		options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Session
	for cur.Next(ctx) {
		var s model.Session
		if err := cur.Decode(&s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, cur.Err()
}
