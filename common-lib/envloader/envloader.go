package envloader

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Constants for environment modes
const (
	Development = "development"
	Production  = "production"
)

// Load loads environment variables from .env file
func Load() {
	// Try loading from the current directory first
	err := godotenv.Load(".env")
	if err != nil {
		// If not found, try the parent directory
		err = godotenv.Load("../.env")
		if err != nil {
			// If not found, try the project root
			err = godotenv.Load("../../.env")
			if err != nil {
				log.Println("Warning: Error loading .env file, using system environment variables")
			}
		}
	}

	// Set defaults if not provided
	if os.Getenv("ENV") == "" {
		os.Setenv("ENV", Development)
	}

	if os.Getenv("CURRENT_RUN") == "" {
		os.Setenv("CURRENT_RUN", Development)
	}
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// IsProduction checks if the current environment is production
func IsProduction() bool {
	return os.Getenv("ENV") == Production
}

// IsDevelopment checks if the current environment is development
func IsDevelopment() bool {
	return os.Getenv("ENV") == Development || os.Getenv("ENV") == ""
}

// GetCurrentRun returns the current running environment
func GetCurrentRun() string {
	currentRun := os.Getenv("CURRENT_RUN")
	if currentRun == "" {
		return Development
	}
	return currentRun
}

// IsCurrentRunValid checks if the current run environment is valid
func IsCurrentRunValid() bool {
	currentRun := GetCurrentRun()
	return currentRun == Development || currentRun == Production
}
