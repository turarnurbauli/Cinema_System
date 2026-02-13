package handler

import (
	"cinema-system/middleware"
	"cinema-system/model"
	"cinema-system/repository"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// ClientHandler возвращает данные клиента (только для admin/cashier).
type ClientHandler struct {
	userRepo    *repository.UserRepo
	bookingRepo *repository.BookingRepo
}

func NewClientHandler(userRepo *repository.UserRepo, bookingRepo *repository.BookingRepo) *ClientHandler {
	return &ClientHandler{userRepo: userRepo, bookingRepo: bookingRepo}
}

func (h *ClientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)
	if role != "admin" && role != "cashier" {
		http.Error(w, `{"error":"forbidden: admin or cashier only"}`, http.StatusForbidden)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/users")
	path = strings.Trim(path, "/")
	if path == "" {
		http.Error(w, `{"error":"user id required"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	u, err := h.userRepo.GetByID(ctx, id)
	if err != nil || u == nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	bookings, _ := h.bookingRepo.GetByUserID(id)
	if bookings == nil {
		bookings = []*model.Booking{}
	}

	type userPublic struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	out := struct {
		User         userPublic       `json:"user"`
		LastBookings []*model.Booking `json:"lastBookings"`
	}{
		User:         userPublic{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role},
		LastBookings: bookings,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
