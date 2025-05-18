package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ServiceRouter routes requests to appropriate microservices
type ServiceRouter struct {
	userServiceURL *url.URL
	// Add more services as needed
}

// NewServiceRouter creates a new service router
func NewServiceRouter(userServiceURL string) (*ServiceRouter, error) {
	userURL, err := url.Parse(userServiceURL)
	if err != nil {
		return nil, err
	}

	return &ServiceRouter{
		userServiceURL: userURL,
	}, nil
}

// Handler returns an http.Handler that routes requests to appropriate services
func (sr *ServiceRouter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the service name from the URL path
		path := r.URL.Path
		parts := strings.Split(path, "/")

		if len(parts) < 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// The service name is the first part of the path after the initial slash
		serviceName := parts[1]

		// Route to the appropriate service based on the service name
		switch serviceName {
		case "users":
			sr.handleUserService(w, r)
		// Add more services here
		default:
			http.Error(w, "Unknown service", http.StatusNotFound)
		}
	})
}

// handleUserService proxies requests to the user service
func (sr *ServiceRouter) handleUserService(w http.ResponseWriter, r *http.Request) {
	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(sr.userServiceURL)

	// Update the request URL
	r.URL.Host = sr.userServiceURL.Host
	r.URL.Scheme = sr.userServiceURL.Scheme

	// Remove the "/users" prefix from the path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/users")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}
