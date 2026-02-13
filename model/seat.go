package model

// Seat — место в зале (ERD: id, hall_id, row_number, seat_number, seat_type).
// seat_type: "regular", "vip"
type Seat struct {
	ID         int    `json:"id" bson:"id"`
	HallID     int    `json:"hallId" bson:"hall_id"`
	RowNumber  int    `json:"rowNumber" bson:"row_number"`   // ряд (1-based)
	SeatNumber int    `json:"seatNumber" bson:"seat_number"` // место в ряду (1-based)
	SeatType   string `json:"seatType" bson:"seat_type"`     // "regular", "vip"
}
