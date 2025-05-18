package utils

import (
	"io/ioutil"
	"os"
	"strings"
)

// IsRunningInDocker phát hiện xem code có đang chạy trong Docker container hay không
func IsRunningInDocker() bool {
	// Cách 1: Kiểm tra biến môi trường
	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return true
	}

	// Cách 2: Kiểm tra file /.dockerenv (chỉ có trong Docker container)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Cách 3: Kiểm tra cgroup
	if data, err := ioutil.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker")
	}

	return false
}

// GetConsulURL trả về URL của Consul dựa trên môi trường
func GetConsulURL() string {
	if IsRunningInDocker() {
		return "consul:8500" // Sử dụng tên service trong Docker
	}
	return "127.0.0.1:8500" // Sử dụng localhost trong môi trường local
}
