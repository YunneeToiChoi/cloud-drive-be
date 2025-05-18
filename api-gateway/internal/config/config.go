package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port           int
	UserServiceURL string
	ConsulURL      string
	Environment    string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:           getEnvAsInt("PORT", 8000),
		UserServiceURL: getEnvAsString("USER_SERVICE_URL", "http://localhost:9001"),
		ConsulURL:      getEnvAsString("CONSUL_URL", "http://localhost:8500"),
		Environment:    getEnvAsString("ENVIRONMENT", "development"),
	}
}

// getEnvAsString returns the environment variable as a string or the default value
func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns the environment variable as an int or the default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
