package clients

import (
	"context"
	"fmt"
	"github.com/cloud-drive/proto-definitions/user"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// UserClient is a client for the user service
type UserClient struct {
	client    user.UserServiceClient
	conn      *grpc.ClientConn
	consulURL string
	serviceID string
}

// UserClientResponse represents a response from the user service
type UserClientResponse struct {
	Users []*user.User `json:"users"`
}

// CreateUserRequest là request cho việc tạo người dùng
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// NewUserClient creates a new user service client
func NewUserClient(consulURL string, fallbackURL string) (*UserClient, error) {
	serviceID := "user-service"

	// Tạo Consul client
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulURL

	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Printf("Failed to create Consul client: %v", err)
		// Nếu không thể kết nối Consul, sử dụng fallbackURL
		return createDirectClient(serviceID, fallbackURL)
	}

	// Tìm service từ Consul
	serviceURL, err := discoverService(consulClient, serviceID)
	if err != nil {
		log.Printf("Failed to discover service: %v, using direct URL", err)
		return createDirectClient(serviceID, fallbackURL)
	}

	// Tạo kết nối gRPC tới service đã tìm thấy
	conn, err := grpc.Dial(serviceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial user service: %v", err)
	}

	client := user.NewUserServiceClient(conn)

	return &UserClient{
		client:    client,
		conn:      conn,
		consulURL: consulURL,
		serviceID: serviceID,
	}, nil
}

// createDirectClient tạo kết nối trực tiếp khi không thể sử dụng Consul
func createDirectClient(serviceID string, fallbackURL string) (*UserClient, error) {
	conn, err := grpc.Dial(fallbackURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial user service directly: %v", err)
	}

	client := user.NewUserServiceClient(conn)

	return &UserClient{
		client:    client,
		conn:      conn,
		serviceID: serviceID,
	}, nil
}

// discoverService tìm service từ Consul
func discoverService(client *consulapi.Client, serviceID string) (string, error) {
	// Tìm service healthy từ Consul
	services, _, err := client.Health().Service(serviceID, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to discover service: %v", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceID)
	}

	// Trả về địa chỉ của service đầu tiên tìm thấy
	service := services[0].Service
	return fmt.Sprintf("%s:%d", service.Address, service.Port), nil
}

// Close closes the client connection
func (c *UserClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ListUsers lists users from the user service
func (c *UserClient) ListUsers(ctx context.Context, limit int, offset int) (*UserClientResponse, error) {
	log.Printf("Calling ListUsers with client: %v", c.client)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &user.ListUsersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	resp, err := c.client.ListUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}

	return &UserClientResponse{
		Users: resp.Users,
	}, nil
}

// CreateUser creates a new user
func (c *UserClient) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.CreateUser(ctx, req)
}

// GetUser gets a user by ID
func (c *UserClient) GetUser(ctx context.Context, id string) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetUser(ctx, &user.GetUserRequest{Id: id})
}

// UpdateUser updates a user
func (c *UserClient) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.UpdateUser(ctx, req)
}

// DeleteUser deletes a user
func (c *UserClient) DeleteUser(ctx context.Context, id string) (*user.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.DeleteUser(ctx, &user.DeleteUserRequest{Id: id})
}

// Authenticate xác thực người dùng
func (c *UserClient) Authenticate(ctx context.Context, email, password string) (*user.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.Authenticate(ctx, &user.AuthRequest{
		Email:    email,
		Password: password,
	})
}
