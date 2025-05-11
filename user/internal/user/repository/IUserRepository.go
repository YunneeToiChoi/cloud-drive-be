package repository

import "user-service/internal/user/model"

type IUserRepository interface {
	GetByID(id string) (*model.User, error)
}
