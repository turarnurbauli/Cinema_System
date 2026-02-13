package model

// Booking — бронирование (ERD: id, user_id, session_id, status, total_price, created_at).
// Status: "pending", "confirmed", "cancelled"
type Booking struct {
	ID         int     `json:"id" bson:"id"`
	UserID     int     `json:"userId" bson:"user_id"`
	SessionID  int     `json:"sessionId" bson:"session_id"`
	Status     string  `json:"status" bson:"status"`
	TotalPrice float64 `json:"totalPrice" bson:"total_price"`
	CreatedAt  string  `json:"createdAt" bson:"created_at"` // RFC3339
	Tickets    []Ticket `json:"tickets,omitempty" bson:"tickets,omitempty"`
}
