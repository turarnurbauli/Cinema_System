package handler

import (
	"cinema-system/model"
	"cinema-system/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler обрабатывает регистрацию и логин.
type AuthHandler struct {
	svc       *service.UserService
	jwtSecret []byte
}

func NewAuthHandler(svc *service.UserService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		svc:       svc,
		jwtSecret: []byte(jwtSecret),
	}
}

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

// Register регистрирует обычного пользователя‑покупателя.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	u, err := h.svc.RegisterCustomer(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	token, err := h.createToken(u)
	if err != nil {
		http.Error(w, `{"error":"failed to create token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&authResponse{
		User:  u,
		Token: token,
	})
}

// Login выполняет вход пользователя и возвращает JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	u, err := h.svc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.createToken(u)
	if err != nil {
		http.Error(w, `{"error":"failed to create token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&authResponse{
		User:  u,
		Token: token,
	})
}

func (h *AuthHandler) createToken(u *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

