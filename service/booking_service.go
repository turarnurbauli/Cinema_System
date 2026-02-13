package service

import (
	"cinema-system/model"
	"cinema-system/repository"
	"context"
	"errors"
	"time"
)

type BookingService struct {
	bookingRepo *repository.BookingRepo
	sessionRepo *repository.SessionRepo
	seatRepo    *repository.SeatRepo
}

func NewBookingService(
	bookingRepo *repository.BookingRepo,
	sessionRepo *repository.SessionRepo,
	seatRepo *repository.SeatRepo,
) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		sessionRepo: sessionRepo,
		seatRepo:    seatRepo,
	}
}

func (s *BookingService) GetByID(id int) (*model.Booking, error) {
	return s.bookingRepo.GetByID(id)
}

func (s *BookingService) GetByUserID(userID int) ([]*model.Booking, error) {
	return s.bookingRepo.GetByUserID(userID)
}

func (s *BookingService) GetAll() ([]*model.Booking, error) {
	return s.bookingRepo.GetAll()
}

// Ticket type prices (same as frontend). VIP seat = 2× adult price.
var ticketTypePrice = map[string]float64{"adult": 2500, "student": 1900, "child": 1600}
const vipPriceMultiplier = 2
const adultPrice = 2500

// Create создаёт бронирование: проверяет сеанс, места, считает цену и сохраняет.
// If ticketTypes has same length as seatIDs, uses Adult/Student/Child prices; else session price + vip 1.5x.
func (s *BookingService) Create(ctx context.Context, userID int, sessionID int, seatIDs []int, ticketTypes []string) (*model.Booking, error) {
	if len(seatIDs) == 0 {
		return nil, errors.New("at least one seat required")
	}
	sess, err := s.sessionRepo.GetByID(sessionID)
	if err != nil || sess == nil {
		return nil, errors.New("session not found")
	}
	booked, err := s.bookingRepo.BookedSeatIDsForSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	bookedSet := make(map[int]bool)
	for _, id := range booked {
		bookedSet[id] = true
	}
	useTicketTypes := len(ticketTypes) == len(seatIDs)
	var tickets []model.Ticket
	var totalPrice float64
	for i, seatID := range seatIDs {
		if bookedSet[seatID] {
			return nil, errors.New("seat already booked")
		}
		seat, err := s.seatRepo.GetByID(seatID)
		if err != nil || seat == nil || seat.HallID != sess.HallID {
			return nil, errors.New("invalid seat")
		}
		var price float64
		if useTicketTypes {
			if seat.SeatType == "vip" {
				price = adultPrice * vipPriceMultiplier
			} else {
				tt := "adult"
				if i < len(ticketTypes) && (ticketTypes[i] == "student" || ticketTypes[i] == "child") {
					tt = ticketTypes[i]
				}
				if p, ok := ticketTypePrice[tt]; ok {
					price = p
				} else {
					price = ticketTypePrice["adult"]
				}
			}
		} else {
			price = sess.Price
			if seat.SeatType == "vip" {
				price = adultPrice * vipPriceMultiplier
			}
		}
		tickets = append(tickets, model.Ticket{SeatID: seatID, Price: price})
		totalPrice += price
		bookedSet[seatID] = true
	}
	b := &model.Booking{
		UserID:     userID,
		SessionID:  sessionID,
		Status:     "confirmed",
		TotalPrice: totalPrice,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Tickets:    tickets,
	}
	return s.bookingRepo.Create(b)
}

func (s *BookingService) Cancel(id int) error {
	return s.bookingRepo.UpdateStatus(id, "cancelled")
}

// ChangeSeats updates a booking's seats (same session). For cashier/admin.
func (s *BookingService) ChangeSeats(ctx context.Context, bookingID int, newSeatIDs []int) (*model.Booking, error) {
	if len(newSeatIDs) == 0 {
		return nil, errors.New("at least one seat required")
	}
	b, err := s.bookingRepo.GetByID(bookingID)
	if err != nil || b == nil {
		return nil, errors.New("booking not found")
	}
	if b.Status == "cancelled" {
		return nil, errors.New("cannot change seats for cancelled booking")
	}
	sess, err := s.sessionRepo.GetByID(b.SessionID)
	if err != nil || sess == nil {
		return nil, errors.New("session not found")
	}
	booked, err := s.bookingRepo.BookedSeatIDsForSession(ctx, b.SessionID)
	if err != nil {
		return nil, err
	}
	currentSeats := make(map[int]bool)
	for _, t := range b.Tickets {
		currentSeats[t.SeatID] = true
	}
	bookedSet := make(map[int]bool)
	for _, id := range booked {
		if !currentSeats[id] {
			bookedSet[id] = true
		}
	}
	var tickets []model.Ticket
	var totalPrice float64
	for _, seatID := range newSeatIDs {
		if bookedSet[seatID] {
			return nil, errors.New("seat already booked")
		}
		seat, err := s.seatRepo.GetByID(seatID)
		if err != nil || seat == nil || seat.HallID != sess.HallID {
			return nil, errors.New("invalid seat")
		}
		price := sess.Price
		if seat.SeatType == "vip" {
			price = adultPrice * vipPriceMultiplier
		}
		tickets = append(tickets, model.Ticket{BookingID: bookingID, SeatID: seatID, Price: price})
		totalPrice += price
		bookedSet[seatID] = true
	}
	if err := s.bookingRepo.UpdateTickets(bookingID, tickets, totalPrice); err != nil {
		return nil, err
	}
	b.Tickets = tickets
	b.TotalPrice = totalPrice
	return b, nil
}
