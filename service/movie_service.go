package service

import (
	"cinema-system/model"
	"cinema-system/repository"
)

// MovieService implements business logic for movies (Assignment 3 Service layer).
type MovieService struct {
	repo *repository.MovieRepo
}

// NewMovieService creates a new movie service.
func NewMovieService(repo *repository.MovieRepo) *MovieService {
	return &MovieService{repo: repo}
}

// Create creates a new movie.
func (s *MovieService) Create(m *model.Movie) (*model.Movie, error) {
	return s.repo.Create(m)
}

// GetByID returns a movie by ID.
func (s *MovieService) GetByID(id int) (*model.Movie, error) {
	return s.repo.GetByID(id)
}

// GetAll returns all movies.
func (s *MovieService) GetAll() ([]*model.Movie, error) {
	return s.repo.GetAll()
}

// Update updates a movie.
func (s *MovieService) Update(m *model.Movie) error {
	return s.repo.Update(m)
}

// Delete deletes a movie by ID.
func (s *MovieService) Delete(id int) error {
	return s.repo.Delete(id)
}
