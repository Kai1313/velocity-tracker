package service_test

import (
	"context"
	"errors"
	"testing"

	"velocity-tracker/backend/internal/apperr"
	"velocity-tracker/backend/internal/model"
	"velocity-tracker/backend/internal/service"
)

type fakeUserRepo struct {
	users  map[int64]*model.User
	nextID int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[int64]*model.User{}}
}

func (f *fakeUserRepo) Create(ctx context.Context, u *model.User) error {
	f.nextID++
	u.ID = f.nextID
	cp := *u
	f.users[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) Get(ctx context.Context, id int64) (*model.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, apperr.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (f *fakeUserRepo) List(ctx context.Context) ([]model.User, error) {
	var out []model.User
	for _, u := range f.users {
		out = append(out, *u)
	}
	return out, nil
}

func (f *fakeUserRepo) Update(ctx context.Context, u *model.User) error {
	if _, ok := f.users[u.ID]; !ok {
		return apperr.ErrNotFound
	}
	cp := *u
	f.users[u.ID] = &cp
	return nil
}

func (f *fakeUserRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := f.users[id]; !ok {
		return apperr.ErrNotFound
	}
	delete(f.users, id)
	return nil
}

func TestUserService_Create_RejectsInvalidRole(t *testing.T) {
	svc := service.NewUserService(newFakeUserRepo())
	_, err := svc.Create(context.Background(), "Alice", model.Role("Manager"))
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("Create() = %v, want apperr.ErrValidation", err)
	}
}

func TestUserService_Create_Valid(t *testing.T) {
	svc := service.NewUserService(newFakeUserRepo())
	u, err := svc.Create(context.Background(), "Alice", model.RoleDeveloper)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if u.Name != "Alice" || u.Role != model.RoleDeveloper {
		t.Fatalf("Create() = %+v, want Alice/Developer", u)
	}
}
