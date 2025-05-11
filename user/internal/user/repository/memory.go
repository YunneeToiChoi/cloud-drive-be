package repository

import (
	"errors"
	"user-service/internal/user/model"
)

type InMemoryUserRepo struct {
	users map[string]*model.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users: map[string]*model.User{
			"1": {ID: "1", Name: "Nguyen Van A", Email: "a@example.com"},
			"2": {ID: "2", Name: "Le Thi B", Email: "b@example.com"},
		},
	}
}

func (r *InMemoryUserRepo) GetByID(id string) (*model.User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}
