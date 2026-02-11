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
			Title:       "Ван Пис фильм: Ред",
			Description: "Музыкальное приключение пиратов Шляпной Команды и дивы Уты.",
			Duration:    115,
			Genre:       "Аниме, Приключения",
			Rating:      8.6,
		},
		{
			Title:       "50 оттенков серого",
			Description: "Романтическая драма о необычном контракте между Анастейшей и Кристианом Греем.",
			Duration:    125,
			Genre:       "Романтика, Драма",
			Rating:      6.1,
		},
		{
			Title:       "Иллюзия обмана",
			Description: "Команда иллюзионистов совершает дерзкие ограбления прямо на сцене.",
			Duration:    115,
			Genre:       "Криминал, Триллер",
			Rating:      7.3,
		},
		{
			Title:       "Один дома",
			Description: "Мальчик, которого забыли дома на Рождество, защищает дом от грабителей.",
			Duration:    103,
			Genre:       "Комедия, Семейный",
			Rating:      8.0,
		},
		{
			Title:       "Demon Slayer: Mugen Train",
			Description: "Тандзиро и отряд охотников на демонов исследуют таинственный поезд.",
			Duration:    118,
			Genre:       "Аниме, Экшен",
			Rating:      8.7,
		},
		{
			Title:       "Бойцовский клуб",
			Description: "Офисный работник создаёт подпольный бойцовский клуб и теряет контроль.",
			Duration:    139,
			Genre:       "Драма, Триллер",
			Rating:      8.8,
		},
		{
			Title:       "Мост в Терабитию",
			Description: "Двое детей создают волшебный мир, чтобы уйти от реальности.",
			Duration:    96,
			Genre:       "Фэнтези, Семейный",
			Rating:      7.2,
		},
		{
			Title:       "Зелёная книга",
			Description: "История дружбы музыканта и его водителя в США 60-х годов.",
			Duration:    130,
			Genre:       "Драма, Биография",
			Rating:      8.2,
		},
		{
			Title:       "Кайтадан",
			Description: "Современная казахстанская драма о выборе и ответственности.",
			Duration:    110,
			Genre:       "Драма",
			Rating:      7.5,
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
