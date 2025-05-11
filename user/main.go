package main

import (
	"common-lib/discovery"
	"common-lib/envloader"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"user-service/internal/user/handler"
	"user-service/internal/user/repository"
	"user-service/internal/user/service"
)

func main() {
	// Load environment variables
	envloader.Load()

	r := gin.Default()

	userRepo := repository.NewUserRepository()

	userService := service.NewUserService(userRepo)

	userHandler := handler.NewUserHandler(userService)

	// Cấu hình routing
	api := r.Group("/api")
	{
		users := api.Group("/users")
		{
			// Health check endpoint
			users.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status": "User service is healthy",
				})
			})

			users.GET("/GetById", userHandler.GetUserByID)
		}
	}

	// Đăng ký service với Consul
	// Lấy thông tin service từ biến môi trường
	serviceName := envloader.GetEnv("SERVICE_NAME", "user")
	servicePortStr := envloader.GetEnv("SERVICE_PORT", "8081")

	servicePort, err := strconv.Atoi(servicePortStr)
	if err != nil {
		servicePort = 8081
	}

	// Tạo URL cho health check
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	healthCheckURL := fmt.Sprintf("http://%s:%d/api/users/health", hostname, servicePort)

	// Khởi tạo service discovery với retry
	log.Println("Initializing Consul service discovery...")
	serviceDiscovery, err := discovery.GetConsulServiceDiscovery()
	if err != nil {
		log.Printf("Warning: Could not connect to Consul: %v", err)
	} else {
		// Đăng ký service với retry
		log.Printf("Registering service '%s' with Consul...", serviceName)
		err = serviceDiscovery.Register(serviceName, servicePort, healthCheckURL)
		if err != nil {
			log.Printf("Warning: Could not register service with Consul: %v", err)
		} else {
			log.Printf("Service '%s' successfully registered with Consul", serviceName)
			// Đảm bảo service được deregister khi shutdown
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				log.Println("Deregistering service and shutting down...")
				serviceDiscovery.Deregister()
				os.Exit(0)
			}()
		}
	}

	log.Println("User service starting on port", servicePort)
	if err := r.Run(":" + servicePortStr); err != nil {
		log.Fatalf("Failed to start User Service: %v", err)
	}
}
