package model

// Session — сеанс (ERD: id, movie_id, hall_id, start_time, price).
// StartTime в формате RFC3339 для API.
type Session struct {
	ID       int     `json:"id" bson:"id"`
	MovieID  int     `json:"movieId" bson:"movie_id"`
	HallID   int     `json:"hallId" bson:"hall_id"`
	StartTime string `json:"startTime" bson:"start_time"` // "2026-02-15T14:00:00Z"
	Price    float64 `json:"price" bson:"price"`          // базовая цена билета
}
