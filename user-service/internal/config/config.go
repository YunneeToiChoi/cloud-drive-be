package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Environment string
	HostMode    string
	Port        int
	ConsulURL   string
	LogLevel    string
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() *Config {
	// Đọc biến môi trường HOST_MODE
	hostMode := getEnv("HOST_MODE", "")

	// Nếu HOST_MODE không được cung cấp, tự động phát hiện
	if hostMode == "" {
		// Kiểm tra các dấu hiệu của Docker container
		if _, err := os.Stat("/.dockerenv"); err == nil {
			// File /.dockerenv tồn tại -> đang trong Docker
			hostMode = "docker"
		} else {
			// Mặc định là local
			hostMode = "local"
		}
	}

	cfg := &Config{
		Environment: getEnv("APP_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HostMode:    hostMode,
	}

	// Lấy cấu hình dựa vào môi trường
	envPrefix := "DEV_"
	if cfg.Environment == "production" {
		envPrefix = "PROD_"
	}

	// Xác định URL Consul dựa trên chế độ host
	if cfg.HostMode == "docker" {
		// Trong Docker, sử dụng tên service làm hostname
		cfg.ConsulURL = getEnv("CONSUL_URL", "consul:8500")
	} else {
		// Khi debug local, sử dụng IP address
		cfg.ConsulURL = getEnv(envPrefix+"CONSUL_URL", "127.0.0.1:8500")
	}

	// Cổng mặc định là 9001 cho user service
	portStr := getEnv("PORT", "9001")
	cfg.Port, _ = strconv.Atoi(portStr)

	// Cấu hình Database - cũng dựa vào HOST_MODE
	if cfg.HostMode == "docker" {
		// Trong Docker, dùng tên service của DB
		cfg.DBHost = getEnv("DB_HOST", "postgres")
	} else {
		// Local thì dùng localhost
		cfg.DBHost = getEnv(envPrefix+"DB_HOST", "localhost")
	}

	dbPortStr := getEnv("DB_PORT", getEnv(envPrefix+"DB_PORT", "5432"))
	cfg.DBPort, _ = strconv.Atoi(dbPortStr)
	cfg.DBUser = getEnv("DB_USER", getEnv(envPrefix+"DB_USER", "postgres"))
	cfg.DBPassword = getEnv("DB_PASSWORD", getEnv(envPrefix+"DB_PASSWORD", "postgres"))
	cfg.DBName = getEnv("DB_NAME", getEnv(envPrefix+"DB_NAME", "users"))

	return cfg
}

// getEnv returns the environment variable or the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
