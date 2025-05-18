package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloud-drive/api-gateway/internal/clients"
	"github.com/cloud-drive/api-gateway/internal/config"
	"github.com/cloud-drive/api-gateway/internal/handlers"
	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Parse command-line flags
	flag.IntVar(&cfg.Port, "port", cfg.Port, "API Gateway port")
	flag.Parse()

	// Create router
	router := mux.NewRouter()

	// Create user service client
	userClient, err := clients.NewUserClient(cfg.ConsulURL, cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("Failed to create user service client: %v", err)
	}
	defer userClient.Close()

	// Create service router
	serviceRouter, err := handlers.NewServiceRouter(cfg.UserServiceURL)
	if err != nil {
		log.Fatalf("Failed to create service router: %v", err)
	}

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Gateway is healthy"))
	})

	// API endpoints for user service
	userRouter := router.PathPrefix("/api/users").Subrouter()

	// Get users
	userRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request for /api/users")
		ctx := r.Context()

		log.Printf("Calling ListUsers with client: %v", userClient)
		resp, err := userClient.ListUsers(ctx, 10, 0)
		if err != nil {
			log.Printf("Error calling ListUsers: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get users: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Got response from ListUsers: %+v", resp)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully responded to /api/users")
	}).Methods("GET")

	// Create user
	userRouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		// Implementation will be added later
	}).Methods("POST")

	// Get user by ID
	userRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Implementation will be added later
	}).Methods("GET")

	// Update user
	userRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Implementation will be added later
	}).Methods("PUT")

	// Delete user
	userRouter.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Implementation will be added later
	}).Methods("DELETE")

	// Legacy proxy routes
	router.PathPrefix("/users").Handler(serviceRouter.Handler())

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Register with Consul
	go registerWithConsul(cfg)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting API Gateway on port %d\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Gateway...")

	// Gracefully shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("API Gateway stopped")
}

// registerWithConsul registers the service with Consul
func registerWithConsul(cfg *config.Config) {
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = cfg.ConsulURL

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Printf("Failed to create Consul client: %v", err)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      fmt.Sprintf("api-gateway-%s-%d", hostname, cfg.Port),
		Name:    "api-gateway",
		Port:    cfg.Port,
		Address: hostname,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", hostname, cfg.Port),
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// Register service
	for i := 0; i < 5; i++ {
		err := client.Agent().ServiceRegister(registration)
		if err == nil {
			log.Printf("Registered service with Consul")
			break
		}
		log.Printf("Failed to register with Consul: %v, retrying...", err)
		time.Sleep(time.Second * 2)
	}

	// Re-register every 30 seconds as a keep-alive mechanism
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		if err := client.Agent().ServiceRegister(registration); err != nil {
			log.Printf("Failed to re-register with Consul: %v", err)
		}
	}
}
