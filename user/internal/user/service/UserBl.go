package service

import (
	"user-service/internal/user/dto"
	"user-service/internal/user/repository"
)

type userServiceImpl struct {
	repo repository.IUserRepository
}

func NewUserService(repo repository.IUserRepository) IUserService {
	return &userServiceImpl{repo: repo}
}

func (s *userServiceImpl) GetUserByID(id string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
