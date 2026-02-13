package model

// Hall — кинозал (ERD: id, name, capacity, seat_layout).
type Hall struct {
	ID         int    `json:"id" bson:"id"`
	Name       string `json:"name" bson:"name"`
	Capacity   int    `json:"capacity" bson:"capacity"`
	Rows       int    `json:"rows" bson:"rows"`             // число рядов
	SeatsPerRow int   `json:"seatsPerRow" bson:"seats_per_row"` // мест в ряду
}
