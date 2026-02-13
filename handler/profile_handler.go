package handler

import (
	"cinema-system/middleware"
	"cinema-system/service"
	"encoding/json"
	"net/http"
)

// ProfileHandler обслуживает профиль текущего авторизованного пользователя.
type ProfileHandler struct {
	userSvc *service.UserService
}

func NewProfileHandler(userSvc *service.UserService) *ProfileHandler {
	return &ProfileHandler{userSvc: userSvc}
}

type profileUpdateRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	AvatarURL       string `json:"avatarUrl"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok || userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		u, err := h.userSvc.GetByID(r.Context(), userID)
		if err != nil || u == nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	case http.MethodPut:
		var req profileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		u, err := h.userSvc.UpdateProfile(r.Context(), userID, service.ProfileUpdate{
			Name:            req.Name,
			Email:           req.Email,
			AvatarURL:       req.AvatarURL,
			CurrentPassword: req.CurrentPassword,
			NewPassword:     req.NewPassword,
		})
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(u)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

