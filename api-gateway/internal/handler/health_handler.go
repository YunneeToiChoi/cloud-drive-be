package handler

import (
	"api-gateway/internal/pkg/client"
	"common-lib/envloader"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var servicesToCheck = []string{"users"}

func HealthCheck(c *gin.Context) {
	currentRun := envloader.GetCurrentRun()

	// Get port based on environment
	var port, host, apiUrl string
	if envloader.IsProduction() {
		port = envloader.GetEnv("PROD_PORT", "80")
		host = envloader.GetEnv("PROD_HOST", "")
		apiUrl = envloader.GetEnv("PROD_API_URL", "")
	} else {
		port = envloader.GetEnv("DEV_PORT", "8080")
		host = envloader.GetEnv("DEV_HOST", "")
		apiUrl = envloader.GetEnv("DEV_API_URL", "")
	}

	// Check if the current run mode is valid
	isValid := envloader.IsCurrentRunValid()

	if isValid {
		c.JSON(http.StatusOK, gin.H{
			"statusCode":  200,
			"status":      "API Gateway is healthy",
			"environment": currentRun,
			"config": gin.H{
				"port":    port,
				"host":    host,
				"api_url": apiUrl,
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode":  500,
			"status":      "API Gateway environment misconfiguration",
			"environment": currentRun,
		})
	}
}

func HealthCheckServices(c *gin.Context) {
	results := map[string]string{}

	// Get current environment
	currentRun := envloader.GetCurrentRun()

	if envloader.IsCurrentRunValid() {
		results["api-gateway"] = "OK"
	} else {
		results["api-gateway"] = "FAIL"
		c.JSON(http.StatusOK, gin.H{
			"statusCode":  200,
			"environment": currentRun,
			"services":    results,
		})
		return
	}

	// Get base URL for other services
	var baseURL string
	if envloader.IsProduction() {
		baseURL = envloader.GetEnv("PROD_API_URL", "")
	} else {
		baseURL = envloader.GetEnv("DEV_API_URL", "")
	}

	// Check base URL availability
	if baseURL == "" {
		results["base-url"] = "URL_NOT_CONFIGURED"
		c.JSON(http.StatusOK, gin.H{
			"statusCode":  200,
			"environment": currentRun,
			"services":    results,
		})
		return
	}

	for i := 0; i < len(servicesToCheck); i++ {
		serviceName := servicesToCheck[i]
		url := baseURL + "/api/" + serviceName + "/health"
		if client.CheckServiceHealth(url) {
			results[serviceName] = "OK"
		} else {
			results[serviceName] = "FAIL"
		}
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	c.JSON(http.StatusOK, gin.H{
		"statusCode":  200,
		"environment": currentRun,
		"services":    results,
	})
}
