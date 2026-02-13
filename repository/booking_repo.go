package repository

import (
	"context"
	"cinema-system/model"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepo struct {
	coll *mongo.Collection
}

func NewBookingRepo(ctx context.Context, client *mongo.Client, dbName string) (*BookingRepo, error) {
	coll := client.Database(dbName).Collection("bookings")
	return &BookingRepo{coll: coll}, nil
}

func (r *BookingRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.Booking
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

func (r *BookingRepo) Create(b *model.Booking) (*model.Booking, error) {
	ctx := context.Background()
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	b.ID = id
	for i := range b.Tickets {
		b.Tickets[i].ID = 0
		b.Tickets[i].BookingID = id
	}
	if _, err := r.coll.InsertOne(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *BookingRepo) GetByID(id int) (*model.Booking, error) {
	ctx := context.Background()
	var b model.Booking
	err := r.coll.FindOne(ctx, bson.D{{Key: "id", Value: id}}).Decode(&b)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepo) GetByUserID(userID int) ([]*model.Booking, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{{Key: "user_id", Value: userID}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Booking
	for cur.Next(ctx) {
		var b model.Booking
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, cur.Err()
}

// GetAll возвращает все бронирования (для админа/кассира).
func (r *BookingRepo) GetAll() ([]*model.Booking, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Booking
	for cur.Next(ctx) {
		var b model.Booking
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, cur.Err()
}

func (r *BookingRepo) GetBySessionID(sessionID int) ([]*model.Booking, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{
		{Key: "session_id", Value: sessionID},
		{Key: "status", Value: bson.D{{Key: "$in", Value: []string{"pending", "confirmed"}}}},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Booking
	for cur.Next(ctx) {
		var b model.Booking
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, cur.Err()
}

func (r *BookingRepo) UpdateStatus(id int, status string) error {
	ctx := context.Background()
	_, err := r.coll.UpdateOne(ctx, bson.D{{Key: "id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: status}}}})
	return err
}

// UpdateTickets replaces tickets and total_price for a booking (for change-seats).
func (r *BookingRepo) UpdateTickets(id int, tickets []model.Ticket, totalPrice float64) error {
	ctx := context.Background()
	ticketsBSON := make(bson.A, 0, len(tickets))
	for _, t := range tickets {
		ticketsBSON = append(ticketsBSON, bson.M{
			"id": t.ID, "booking_id": t.BookingID, "seat_id": t.SeatID, "price": t.Price,
		})
	}
	_, err := r.coll.UpdateOne(ctx, bson.D{{Key: "id", Value: id}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "tickets", Value: ticketsBSON},
			{Key: "total_price", Value: totalPrice},
		}}})
	return err
}

// BookedSeatIDsForSession возвращает ID мест, уже занятых по данному сеансу.
func (r *BookingRepo) BookedSeatIDsForSession(ctx context.Context, sessionID int) ([]int, error) {
	bookings, err := r.GetBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	var seatIDs []int
	seen := make(map[int]bool)
	for _, b := range bookings {
		for _, t := range b.Tickets {
			if !seen[t.SeatID] {
				seen[t.SeatID] = true
				seatIDs = append(seatIDs, t.SeatID)
			}
		}
	}
	return seatIDs, nil
}
