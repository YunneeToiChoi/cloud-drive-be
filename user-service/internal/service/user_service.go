package service

import (
	"context"
	"github.com/cloud-drive/proto-definitions/user"
	"github.com/cloud-drive/user-service/internal/models"
	"github.com/cloud-drive/user-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// UserService implements the gRPC UserService
type UserService struct {
	user.UnimplementedUserServiceServer
	repo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Create user model
	now := time.Now()
	userModel := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user
	if err := s.repo.Create(ctx, userModel); err != nil {
		if err == repository.ErrUserExists {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &user.UserResponse{
		User: convertUserToProto(userModel),
	}, nil
}

// GetUser gets a user by ID
func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	userModel, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &user.UserResponse{
		User: convertUserToProto(userModel),
	}, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
	// Get existing user
	userModel, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Update fields
	if req.Email != "" {
		userModel.Email = req.Email
	}
	if req.FirstName != "" {
		userModel.FirstName = req.FirstName
	}
	if req.LastName != "" {
		userModel.LastName = req.LastName
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
		}
		userModel.Password = string(hashedPassword)
	}
	userModel.UpdatedAt = time.Now()

	// Save user
	if err := s.repo.Update(ctx, userModel); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &user.UserResponse{
		User: convertUserToProto(userModel),
	}, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	if err := s.repo.Delete(ctx, req.Id); err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &user.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers lists users
func (s *UserService) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	users, err := s.repo.List(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// Convert users to proto
	protoUsers := make([]*user.User, 0, len(users))
	for _, u := range users {
		protoUsers = append(protoUsers, convertUserToProto(u))
	}

	return &user.ListUsersResponse{
		Users: protoUsers,
	}, nil
}

// convertUserToProto converts a user model to a proto user
func convertUserToProto(userModel *models.User) *user.User {
	return &user.User{
		Id:        userModel.ID,
		Username:  userModel.Username,
		Email:     userModel.Email,
		FirstName: userModel.FirstName,
		LastName:  userModel.LastName,
		CreatedAt: userModel.CreatedAt.Format(time.RFC3339),
		UpdatedAt: userModel.UpdatedAt.Format(time.RFC3339),
	}
}
