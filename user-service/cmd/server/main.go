package main

import (
	"flag"
	"fmt"
	"github.com/cloud-drive/proto-definitions/user"
	"github.com/cloud-drive/user-service/internal/config"
	"github.com/cloud-drive/user-service/internal/repository"
	"github.com/cloud-drive/user-service/internal/service"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Parse command-line flags (chỉ ghi đè nếu được cung cấp)
	flag.IntVar(&cfg.Port, "port", cfg.Port, "User service gRPC port")
	flag.Parse()

	// Log thông tin môi trường
	log.Printf("Starting User Service in %s environment with host mode: %s", cfg.Environment, cfg.HostMode)

	// Xác định địa chỉ lắng nghe - Quan trọng: sử dụng 0.0.0.0 để các container khác có thể kết nối
	listenAddr := "0.0.0.0"
	if cfg.HostMode == "local" && cfg.Environment == "development" {
		// Khi debug local, có thể dùng localhost hoặc 0.0.0.0
		listenAddr = "0.0.0.0" // Vẫn dùng 0.0.0.0 để các service Docker khác kết nối được
	}

	// Set up listener
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenAddr, cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	server := grpc.NewServer()

	// Create repository
	userRepo := repository.NewInMemoryUserRepository()

	// Create and register user service
	userService := service.NewUserService(userRepo)
	user.RegisterUserServiceServer(server, userService)

	// Register health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	// Register reflection service
	reflection.Register(server)

	// Register service with Consul
	go registerWithConsul(cfg)

	// Start gRPC server
	go func() {
		log.Printf("Starting User Service on %s:%d\n", listenAddr, cfg.Port)
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down User Service...")

	// Gracefully stop server
	server.GracefulStop()
	log.Println("User Service stopped")
}

// registerWithConsul registers the service with Consul
func registerWithConsul(cfg *config.Config) {
	consulConfig := consulapi.DefaultConfig()

	// Sửa URL Consul để sử dụng IPv4
	consulConfig.Address = strings.Replace(cfg.ConsulURL, "localhost", "127.0.0.1", 1)

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Printf("Failed to create Consul client: %v", err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Xác định địa chỉ đăng ký với Consul
	serviceAddress := hostname
	if cfg.HostMode == "local" {
		// Khi chạy debug local, đăng ký với Consul bằng localhost
		serviceAddress = "localhost"
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      fmt.Sprintf("user-service-%s-%d", hostname, cfg.Port),
		Name:    "user-service",
		Port:    cfg.Port,
		Address: serviceAddress,
		Check: &consulapi.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%s:%d", serviceAddress, cfg.Port),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// Initial registration
	retryRegister(client, registration, 5)

	// Re-register every 30 seconds as a keep-alive mechanism
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		if err := client.Agent().ServiceRegister(registration); err != nil {
			log.Printf("Failed to re-register with Consul: %v", err)
		}
	}
}

// retryRegister tries to register with Consul with retries
func retryRegister(client *consulapi.Client, registration *consulapi.AgentServiceRegistration, retries int) {
	for i := 0; i < retries; i++ {
		err := client.Agent().ServiceRegister(registration)
		if err == nil {
			log.Printf("Registered service with Consul")
			return
		}
		log.Printf("Failed to register with Consul: %v, retrying...", err)
		time.Sleep(time.Second * 2)
	}
	log.Printf("Could not register with Consul after %d retries", retries)
}
