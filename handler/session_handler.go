package handler

import (
	"cinema-system/model"
	"cinema-system/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions")
	path = strings.Trim(path, "/")

	if path == "" {
		if r.Method == http.MethodGet {
			h.list(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.create(w, r)
			return
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.SplitN(path, "/", 2)
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if len(parts) == 2 && parts[1] == "seats" {
		if r.Method == http.MethodGet {
			h.seats(w, r, id)
			return
		}
	}
	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			h.getByID(w, r, id)
			return
		}
		if r.Method == http.MethodPut {
			h.update(w, r, id)
			return
		}
	}
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (h *SessionHandler) list(w http.ResponseWriter, r *http.Request) {
	movieIDStr := r.URL.Query().Get("movieId")
	if movieIDStr != "" {
		movieID, err := strconv.Atoi(movieIDStr)
		if err != nil {
			http.Error(w, `{"error":"invalid movieId"}`, http.StatusBadRequest)
			return
		}
		sessions, err := h.svc.GetByMovieID(movieID)
		if err != nil {
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
			return
		}
		if sessions == nil {
			sessions = []*model.Session{}
		}
		_ = json.NewEncoder(w).Encode(sessions)
		return
	}
	sessions, err := h.svc.GetAll()
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if sessions == nil {
		sessions = []*model.Session{}
	}
	_ = json.NewEncoder(w).Encode(sessions)
}

func (h *SessionHandler) getByID(w http.ResponseWriter, _ *http.Request, id int) {
	sess, err := h.svc.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if sess == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(sess)
}

func (h *SessionHandler) seats(w http.ResponseWriter, r *http.Request, sessionID int) {
	seats, bookedIDs, err := h.svc.GetAvailableSeats(r.Context(), sessionID)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	bookedSet := make(map[int]bool)
	for _, id := range bookedIDs {
		bookedSet[id] = true
	}
	type seatWithStatus struct {
		ID         int     `json:"id"`
		HallID     int     `json:"hallId"`
		RowNumber  int     `json:"rowNumber"`
		SeatNumber int     `json:"seatNumber"`
		SeatType   string  `json:"seatType"`
		Booked     bool    `json:"booked"`
	}
	out := make([]seatWithStatus, 0, len(seats))
	for _, s := range seats {
		out = append(out, seatWithStatus{
			ID:         s.ID,
			HallID:     s.HallID,
			RowNumber:  s.RowNumber,
			SeatNumber: s.SeatNumber,
			SeatType:   s.SeatType,
			Booked:     bookedSet[s.ID],
		})
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (h *SessionHandler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MovieID   int     `json:"movieId"`
		HallID    int     `json:"hallId"`
		StartTime string  `json:"startTime"`
		Price     float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	sess, err := h.svc.Create(&model.Session{
		MovieID:   body.MovieID,
		HallID:    body.HallID,
		StartTime: body.StartTime,
		Price:     body.Price,
	})
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sess)
}

func (h *SessionHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var body struct {
		MovieID   int     `json:"movieId"`
		HallID    int     `json:"hallId"`
		StartTime string  `json:"startTime"`
		Price     float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	sess, err := h.svc.GetByID(id)
	if err != nil || sess == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	sess.MovieID = body.MovieID
	sess.HallID = body.HallID
	sess.StartTime = body.StartTime
	sess.Price = body.Price
	if err := h.svc.Update(sess); err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(sess)
}
