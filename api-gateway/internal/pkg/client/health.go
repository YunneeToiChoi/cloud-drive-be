package client

import (
	"net/http"
	"time"
)

func CheckServiceHealth(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	res, err := client.Get(url)
	if err != nil || res.StatusCode != 200 {
		return false
	}
	return true
}
