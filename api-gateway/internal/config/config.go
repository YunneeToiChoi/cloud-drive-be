package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Environment string
	Port        int
	ConsulURL   string
	HostMode    string
	LogLevel    string
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
		// Trong Docker, sử dụng tên service
		cfg.ConsulURL = getEnv("CONSUL_URL", "consul:8500")
	} else {
		// Trong local, sử dụng IPv4 (127.0.0.1)
		cfg.ConsulURL = getEnv(envPrefix+"CONSUL_URL", "127.0.0.1:8500")
	}

	// Service port
	portStr := getEnv("PORT", "")
	if portStr == "" {
		portStr = getEnv(envPrefix+"API_GATEWAY_PORT", "8080")
	}
	cfg.Port, _ = strconv.Atoi(portStr)

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
