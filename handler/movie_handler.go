package handler

import (
	"cinema-system/model"
	"cinema-system/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// MovieHandler handles HTTP requests for movies (Assignment 3 Handlers layer).
type MovieHandler struct {
	svc *service.MovieService
}

// NewMovieHandler creates a new movie HTTP handler.
func NewMovieHandler(svc *service.MovieService) *MovieHandler {
	return &MovieHandler{svc: svc}
}

// ServeHTTP routes requests to list, get, create, update, delete.
func (h *MovieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/movies")
	path = strings.Trim(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.list(w, r)
			return
		case http.MethodPost:
			h.create(w, r)
			return
		}
	}

	// path is ":id"
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getByID(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *MovieHandler) list(w http.ResponseWriter, _ *http.Request) {
	movies, err := h.svc.GetAll()
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if movies == nil {
		movies = []*model.Movie{}
	}
	_ = json.NewEncoder(w).Encode(movies)
}

func (h *MovieHandler) getByID(w http.ResponseWriter, _ *http.Request, id int) {
	m, err := h.svc.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(m)
}

func (h *MovieHandler) create(w http.ResponseWriter, r *http.Request) {
	var m model.Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	created, err := h.svc.Create(&m)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *MovieHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var m model.Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	m.ID = id
	if err := h.svc.Update(&m); err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(&m)
}

func (h *MovieHandler) delete(w http.ResponseWriter, _ *http.Request, id int) {
	_ = h.svc.Delete(id)
	w.WriteHeader(http.StatusNoContent)
}
