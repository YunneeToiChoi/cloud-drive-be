package clients

import (
	"context"
	"fmt"
	pb "github.com/cloud-drive/proto-definitions/user"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// UserClient is a client for the User Service
type UserClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

// NewUserClient creates a new User Service client
func NewUserClient(consulURL, userServiceURL string) (*UserClient, error) {
	// Try to discover user service from Consul first
	serviceURL, err := discoverService(consulURL, "user-service")
	if err != nil {
		log.Printf("Failed to discover user service from Consul: %v, using direct URL", err)
		serviceURL = userServiceURL
	}

	// Connect to user service
	conn, err := grpc.Dial(serviceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %v", err)
	}

	return &UserClient{
		client: pb.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close closes the client connection
func (c *UserClient) Close() error {
	return c.conn.Close()
}

// CreateUser creates a new user
func (c *UserClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.CreateUser(ctx, req)
}

// GetUser gets a user by ID
func (c *UserClient) GetUser(ctx context.Context, id string) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetUser(ctx, &pb.GetUserRequest{Id: id})
}

// UpdateUser updates a user
func (c *UserClient) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.UpdateUser(ctx, req)
}

// DeleteUser deletes a user
func (c *UserClient) DeleteUser(ctx context.Context, id string) (*pb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
}

// ListUsers lists users
func (c *UserClient) ListUsers(ctx context.Context, limit, offset int32) (*pb.ListUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.ListUsers(ctx, &pb.ListUsersRequest{Limit: limit, Offset: offset})
}

// discoverService discovers a service from Consul
func discoverService(consulURL, serviceName string) (string, error) {
	config := consulapi.DefaultConfig()
	config.Address = consulURL

	client, err := consulapi.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Consul client: %v", err)
	}

	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to discover service: %v", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	service := services[0]
	return fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port), nil
}
