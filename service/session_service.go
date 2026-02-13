package service

import (
	"cinema-system/model"
	"cinema-system/repository"
	"context"
)

type SessionService struct {
	sessionRepo *repository.SessionRepo
	hallRepo    *repository.HallRepo
	seatRepo    *repository.SeatRepo
	bookingRepo *repository.BookingRepo
}

func NewSessionService(
	sessionRepo *repository.SessionRepo,
	hallRepo *repository.HallRepo,
	seatRepo *repository.SeatRepo,
	bookingRepo *repository.BookingRepo,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		hallRepo:    hallRepo,
		seatRepo:    seatRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *SessionService) GetByID(id int) (*model.Session, error) {
	return s.sessionRepo.GetByID(id)
}

func (s *SessionService) GetAll() ([]*model.Session, error) {
	return s.sessionRepo.GetAll()
}

func (s *SessionService) GetByMovieID(movieID int) ([]*model.Session, error) {
	return s.sessionRepo.GetByMovieID(movieID)
}

func (s *SessionService) Create(sess *model.Session) (*model.Session, error) {
	return s.sessionRepo.Create(sess)
}

func (s *SessionService) Update(sess *model.Session) error {
	return s.sessionRepo.Update(sess)
}

// GetAvailableSeats возвращает места зала для сеанса; занятые помечены.
func (s *SessionService) GetAvailableSeats(ctx context.Context, sessionID int) ([]*model.Seat, []int, error) {
	sess, err := s.sessionRepo.GetByID(sessionID)
	if err != nil || sess == nil {
		return nil, nil, err
	}
	hall, err := s.hallRepo.GetByID(sess.HallID)
	if err != nil || hall == nil {
		return nil, nil, err
	}
	if err := s.seatRepo.EnsureSeatsForHall(ctx, hall); err != nil {
		return nil, nil, err
	}
	seats, err := s.seatRepo.GetByHallID(hall.ID)
	if err != nil {
		return nil, nil, err
	}
	booked, err := s.bookingRepo.BookedSeatIDsForSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	bookedSet := make(map[int]bool)
	for _, id := range booked {
		bookedSet[id] = true
	}
	return seats, booked, nil
}
