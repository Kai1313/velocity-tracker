package service

import (
	"context"
	"fmt"
	"strings"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	Get(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id int64) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func validateUserFields(name string, role model.Role) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("%w: name is required", apperr.ErrValidation)
	}
	switch role {
	case model.RoleLead, model.RoleDeveloper:
	default:
		return "", fmt.Errorf("%w: role must be Lead or Developer", apperr.ErrValidation)
	}
	return name, nil
}

func (s *UserService) Create(ctx context.Context, name string, role model.Role) (*model.User, error) {
	name, err := validateUserFields(name, role)
	if err != nil {
		return nil, err
	}
	u := &model.User{Name: name, Role: role}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.Get(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id int64, name string, role model.Role) (*model.User, error) {
	name, err := validateUserFields(name, role)
	if err != nil {
		return nil, err
	}
	u := &model.User{ID: id, Name: name, Role: role}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
