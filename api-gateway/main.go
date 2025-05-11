package main

import (
	"api-gateway/internal/router"
	"common-lib/discovery"
	"common-lib/envloader"
	"log"
)

func main() {
	// Load environment variables
	envloader.Load()

	// Khởi tạo kết nối với Consul
	log.Println("Initializing Consul service discovery...")
	_, err := discovery.GetConsulServiceDiscovery()

	// Nếu kết nối thất bại, ghi log và tiếp tục hoạt động (fallback mechanism)
	if err != nil {
		log.Printf("Warning: Could not connect to Consul: %v. Service discovery will use fallback mechanism.", err)
	} else {
		log.Println("Successfully connected to Consul service registry")
	}

	// Khởi tạo router
	r := router.SetupRouter()
	port := envloader.GetEnv("DEV_PORT", "8080")

	log.Printf("API Gateway starting in %s mode on port %s", envloader.GetCurrentRun(), port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
