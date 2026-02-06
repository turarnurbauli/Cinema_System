package model

// Movie matches ERD from Assignment 3: id, title, description, duration, genre, rating.
type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"` // minutes
	Genre       string  `json:"genre"`
	Rating      float64 `json:"rating"`
}
