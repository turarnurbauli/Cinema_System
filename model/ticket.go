package model

// Ticket — билет (ERD: id, booking_id, seat_id, price).
type Ticket struct {
	ID        int     `json:"id" bson:"id"`
	BookingID int     `json:"bookingId" bson:"booking_id"`
	SeatID    int     `json:"seatId" bson:"seat_id"`
	Price     float64 `json:"price" bson:"price"`
}
