package repository

import (
	"context"
	"cinema-system/model"

	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MovieRepo — репозиторий фильмов на MongoDB.
type MovieRepo struct {
	coll *mongo.Collection
}

// NewMovieRepo принимает MongoDB‑клиент и имя базы, инициализирует коллекцию и
// при необходимости выполняет начальное заполнение.
func NewMovieRepo(ctx context.Context, client *mongo.Client, dbName string) (*MovieRepo, error) {
	coll := client.Database(dbName).Collection("movies")
	r := &MovieRepo{coll: coll}
	if err := r.seedIfEmpty(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *MovieRepo) seedIfEmpty(ctx context.Context) error {
	count, err := r.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// Seed with a few example movies so UI and API are not empty on first run.
	seed := []model.Movie{
		{
			Title:       "Inception",
			Description: "A thief who steals corporate secrets through dream-sharing technology.",
			Duration:    148,
			Genre:       "Sci-Fi",
			Rating:      8.8,
		},
		{
			Title:       "The Dark Knight",
			Description: "Batman faces the Joker in Gotham City.",
			Duration:    152,
			Genre:       "Action",
			Rating:      9.0,
		},
		{
			Title:       "Interstellar",
			Description: "Explorers travel through a wormhole in space to ensure humanity's survival.",
			Duration:    169,
			Genre:       "Sci-Fi",
			Rating:      8.6,
		},
	}

	for _, m := range seed {
		if _, err := r.Create(&m); err != nil {
			return err
		}
	}

	return nil
}

// nextID ищет максимальный id и возвращает следующий.
func (r *MovieRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.Movie
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

// Create saves a new movie and returns it with ID set.
func (r *MovieRepo) Create(m *model.Movie) (*model.Movie, error) {
	ctx := context.Background()
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	m.ID = id
	if _, err := r.coll.InsertOne(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetByID returns a movie by ID or nil if not found.
func (r *MovieRepo) GetByID(id int) (*model.Movie, error) {
	ctx := context.Background()
	var m model.Movie
	err := r.coll.FindOne(ctx, bson.D{{Key: "id", Value: id}}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetAll returns all movies.
func (r *MovieRepo) GetAll() ([]*model.Movie, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []*model.Movie
	for cur.Next(ctx) {
		var m model.Movie
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Update updates an existing movie by ID.
func (r *MovieRepo) Update(m *model.Movie) error {
	ctx := context.Background()
	update := bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "title", Value: m.Title},
			{Key: "description", Value: m.Description},
			{Key: "duration", Value: m.Duration},
			{Key: "genre", Value: m.Genre},
			{Key: "rating", Value: m.Rating},
		},
	}}
	_, err := r.coll.UpdateOne(ctx, bson.D{{Key: "id", Value: m.ID}}, update)
	return err
}

// Delete removes a movie by ID.
func (r *MovieRepo) Delete(id int) error {
	ctx := context.Background()
	_, err := r.coll.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	return err
}
