package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port        int
	ConsulURL   string
	Environment string
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Port:        getEnvAsInt("PORT", 9001),
		ConsulURL:   getEnvAsString("CONSUL_URL", "http://localhost:8500"),
		Environment: getEnvAsString("ENVIRONMENT", "development"),
		DBHost:      getEnvAsString("DB_HOST", "localhost"),
		DBPort:      getEnvAsInt("DB_PORT", 5432),
		DBUser:      getEnvAsString("DB_USER", "postgres"),
		DBPassword:  getEnvAsString("DB_PASSWORD", "postgres"),
		DBName:      getEnvAsString("DB_NAME", "users"),
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
