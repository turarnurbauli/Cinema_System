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

// NewMovieRepo creates a new in-memory movie repository.
func NewMovieRepo() *MovieRepo {
	return &MovieRepo{
		items:  make(map[int]*model.Movie),
		nextID: 1,
	}
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
