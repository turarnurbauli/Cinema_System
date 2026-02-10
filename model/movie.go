package model

// Movie matches ERD from Assignment 3: id, title, description, duration, genre, rating.
// BSON теги нужны для сохранения в MongoDB.
type Movie struct {
	ID          int     `json:"id" bson:"id"`
	Title       string  `json:"title" bson:"title"`
	Description string  `json:"description" bson:"description"`
	Duration    int     `json:"duration" bson:"duration"` // minutes
	Genre       string  `json:"genre" bson:"genre"`
	Rating      float64 `json:"rating" bson:"rating"`
}
