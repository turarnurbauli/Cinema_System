package repository

import (
	"cinema-system/model"
	"sync"
)

// MovieRepo is an in-memory store with mutex for safe concurrent access.
type MovieRepo struct {
	mu     sync.RWMutex
	items  map[int]*model.Movie
	nextID int
}

// NewMovieRepo creates a new in-memory movie repository and seeds a few demo movies.
func NewMovieRepo() *MovieRepo {
	r := &MovieRepo{
		items:  make(map[int]*model.Movie),
		nextID: 1,
	}

	// Seed with a few example movies so UI and API are not empty on first run.
	seed := []*model.Movie{
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
		_, _ = r.Create(m)
	}

	return r
}

// Create saves a new movie and returns it with ID set.
func (r *MovieRepo) Create(m *model.Movie) (*model.Movie, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m.ID = r.nextID
	r.nextID++
	r.items[m.ID] = m
	return m, nil
}

// GetByID returns a movie by ID or nil if not found.
func (r *MovieRepo) GetByID(id int) (*model.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.items[id]
	if !ok {
		return nil, nil
	}
	return m, nil
}

// GetAll returns all movies.
func (r *MovieRepo) GetAll() ([]*model.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Movie, 0, len(r.items))
	for _, m := range r.items {
		out = append(out, m)
	}
	return out, nil
}

// Update updates an existing movie by ID.
func (r *MovieRepo) Update(m *model.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[m.ID]; !ok {
		return nil // not found, no error for simplicity
	}
	r.items[m.ID] = m
	return nil
}

// Delete removes a movie by ID.
func (r *MovieRepo) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
	return nil
}
