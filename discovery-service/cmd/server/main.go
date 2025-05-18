package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloud-drive/discovery-service/internal/config"
	"github.com/cloud-drive/discovery-service/internal/consul"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Parse command-line flags
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Discovery service HTTP port")
	flag.Parse()

	// Create router
	router := mux.NewRouter()

	// Create Consul client
	consulClient, err := consul.NewConsulClient(cfg.ConsulAddress)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Discovery Service is healthy"))
	})

	// Service discovery endpoints
	router.HandleFunc("/services/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serviceName := vars["name"]

		services, err := consulClient.GetService(serviceName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get service: %v", err), http.StatusInternalServerError)
			return
		}

		if len(services) == 0 {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		// Return the first healthy service instance
		service := services[0]
		addr := fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port)
		w.Write([]byte(addr))
	}).Methods("GET")

	// Service registration endpoint
	router.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name    string   `json:"name"`
			Address string   `json:"address"`
			Port    int      `json:"port"`
			Tags    []string `json:"tags"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
			return
		}

		if err := consulClient.Register(req.Name, req.Address, req.Port, req.Tags); err != nil {
			http.Error(w, fmt.Sprintf("Failed to register service: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service registered"))
	}).Methods("POST")

	// Service deregistration endpoint
	router.HandleFunc("/services/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		serviceID := vars["id"]

		if err := consulClient.Deregister(serviceID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to deregister service: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Service deregistered"))
	}).Methods("DELETE")

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Discovery Service on port %d\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Discovery Service...")

	// Gracefully shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Discovery Service stopped")
}
