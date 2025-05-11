package service

import "user-service/internal/user/dto"

type IUserService interface {
	GetUserByID(id string) (*dto.UserResponse, error)
}
