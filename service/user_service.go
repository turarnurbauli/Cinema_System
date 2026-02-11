package service

import (
	"cinema-system/model"
	"cinema-system/repository"
	"context"

	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Предопределённые роли.
const (
	RoleCustomer = "customer"
	RoleCashier  = "cashier"
	RoleAdmin    = "admin"
)

// UserService инкапсулирует бизнес-логику работы с пользователями.
type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(repo *repository.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// EnsureUserWithRole создаёт пользователя с заданной ролью, если такого email ещё нет.
func (s *UserService) EnsureUserWithRole(ctx context.Context, email, password, name, role string) error {
	if email == "" || password == "" {
		return nil
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.repo.Create(ctx, &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
	})
	return err
}

// EnsureDefaultAdmin создаёт админа, если его ещё нет.
func (s *UserService) EnsureDefaultAdmin(ctx context.Context, email, password, name string) error {
	return s.EnsureUserWithRole(ctx, email, password, name, RoleAdmin)
}

// RegisterCustomer регистрирует обычного пользователя‑покупателя.
func (s *UserService) RegisterCustomer(ctx context.Context, name, email, password string) (*model.User, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         RoleCustomer,
	})
}

// Authenticate проверяет email+пароль и возвращает пользователя.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

