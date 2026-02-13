package handler

import (
	"cinema-system/middleware"
	"cinema-system/model"
	"cinema-system/repository"
	"cinema-system/service"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type BookingHandler struct {
	svc      *service.BookingService
	userRepo *repository.UserRepo
}

func NewBookingHandler(svc *service.BookingService, userRepo *repository.UserRepo) *BookingHandler {
	return &BookingHandler{svc: svc, userRepo: userRepo}
}

// bookingWithUser — ответ с именем клиента.
type bookingWithUser struct {
	*model.Booking
	UserName string `json:"userName"`
}

func (h *BookingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api/bookings")
	path = strings.Trim(path, "/")

	if path == "" {
		if r.Method == http.MethodGet {
			h.listMy(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.create(w, r)
			return
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet {
		h.getByID(w, r, id)
		return
	}
	if r.Method == http.MethodDelete {
		h.cancel(w, r, id)
		return
	}
	if r.Method == http.MethodPatch {
		h.changeSeats(w, r, id)
		return
	}
	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (h *BookingHandler) listMy(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	ctx := r.Context()
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	var list []*model.Booking
	var err error
	if r.URL.Query().Get("all") == "1" && (role == "admin" || role == "cashier") {
		list, err = h.svc.GetAll()
	} else {
		list, err = h.svc.GetByUserID(userID)
	}
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*model.Booking{}
	}
	out := h.enrichWithUserName(ctx, list)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *BookingHandler) enrichWithUserName(ctx context.Context, list []*model.Booking) []bookingWithUser {
	out := make([]bookingWithUser, 0, len(list))
	for _, b := range list {
		name := ""
		if u, _ := h.userRepo.GetByID(ctx, b.UserID); u != nil {
			name = u.Name
		}
		out = append(out, bookingWithUser{Booking: b, UserName: name})
	}
	return out
}

func (h *BookingHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	b, err := h.svc.GetByID(id)
	if err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	if b == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	if ok && b.UserID != userID && role != "admin" && role != "cashier" {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	_ = json.NewEncoder(w).Encode(b)
}

func (h *BookingHandler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	if role == "admin" {
		http.Error(w, `{"error":"admin cannot create bookings"}`, http.StatusForbidden)
		return
	}
	var body struct {
		SessionID   int      `json:"sessionId"`
		SeatIDs     []int    `json:"seatIds"`
		TicketTypes []string `json:"ticketTypes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	b, err := h.svc.Create(r.Context(), userID, body.SessionID, body.SeatIDs, body.TicketTypes)
	if err != nil {
		if err.Error() == "session not found" || err.Error() == "invalid seat" || err.Error() == "seat already booked" || err.Error() == "at least one seat required" {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(b)
}

func (h *BookingHandler) cancel(w http.ResponseWriter, r *http.Request, id int) {
	b, err := h.svc.GetByID(id)
	if err != nil || b == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	if !ok || (b.UserID != userID && role != "admin" && role != "cashier") {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	if err := h.svc.Cancel(id); err != nil {
		http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookingHandler) changeSeats(w http.ResponseWriter, r *http.Request, id int) {
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	if role != "admin" && role != "cashier" {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}
	var body struct {
		SeatIDs []int `json:"seatIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	b, err := h.svc.ChangeSeats(r.Context(), id, body.SeatIDs)
	if err != nil {
		switch err.Error() {
		case "booking not found", "session not found":
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		case "at least one seat required", "invalid seat", "seat already booked", "cannot change seats for cancelled booking":
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error":"internal"}`, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(b)
}
