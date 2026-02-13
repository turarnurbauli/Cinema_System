package repository

import (
	"context"
	"cinema-system/model"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SeatRepo struct {
	coll *mongo.Collection
}

func NewSeatRepo(ctx context.Context, client *mongo.Client, dbName string) (*SeatRepo, error) {
	coll := client.Database(dbName).Collection("seats")
	r := &SeatRepo{coll: coll}
	return r, nil
}

func (r *SeatRepo) nextID(ctx context.Context) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "id", Value: -1}})
	var last model.Seat
	err := r.coll.FindOne(ctx, bson.D{}, opts).Decode(&last)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return last.ID + 1, nil
}

func (r *SeatRepo) Create(s *model.Seat) (*model.Seat, error) {
	ctx := context.Background()
	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}
	s.ID = id
	if _, err := r.coll.InsertOne(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SeatRepo) GetByID(id int) (*model.Seat, error) {
	ctx := context.Background()
	var s model.Seat
	err := r.coll.FindOne(ctx, bson.D{{Key: "id", Value: id}}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SeatRepo) GetByHallID(hallID int) ([]*model.Seat, error) {
	ctx := context.Background()
	cur, err := r.coll.Find(ctx, bson.D{{Key: "hall_id", Value: hallID}},
		options.Find().SetSort(bson.D{{Key: "row_number", Value: 1}, {Key: "seat_number", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Seat
	for cur.Next(ctx) {
		var s model.Seat
		if err := cur.Decode(&s); err != nil {
			return nil, err
		}
		out = append(out, &s)
	}
	return out, cur.Err()
}

// EnsureSeatsForHall создаёт места для зала, если их ещё нет.
func (r *SeatRepo) EnsureSeatsForHall(ctx context.Context, hall *model.Hall) error {
	count, err := r.coll.CountDocuments(ctx, bson.D{{Key: "hall_id", Value: hall.ID}})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for row := 1; row <= hall.Rows; row++ {
		for num := 1; num <= hall.SeatsPerRow; num++ {
			st := "regular"
			if hall.Rows >= 2 && row >= hall.Rows-1 {
				st = "vip"
			}
			_, err := r.Create(&model.Seat{HallID: hall.ID, RowNumber: row, SeatNumber: num, SeatType: st})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// FixVipSeatsForHall sets seat_type to "vip" only for the last 2 rows, "regular" for the rest.
// Call after EnsureSeatsForHall so existing DB data matches the rule (VIP in last 2 rows).
func (r *SeatRepo) FixVipSeatsForHall(ctx context.Context, hall *model.Hall) error {
	if hall.Rows < 2 {
		return nil
	}
	lastTwoStart := hall.Rows - 1
	_, err := r.coll.UpdateMany(ctx,
		bson.D{{Key: "hall_id", Value: hall.ID}, {Key: "row_number", Value: bson.D{{Key: "$gte", Value: lastTwoStart}}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "seat_type", Value: "vip"}}}})
	if err != nil {
		return err
	}
	_, err = r.coll.UpdateMany(ctx,
		bson.D{{Key: "hall_id", Value: hall.ID}, {Key: "row_number", Value: bson.D{{Key: "$lt", Value: lastTwoStart}}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "seat_type", Value: "regular"}}}})
	return err
}
