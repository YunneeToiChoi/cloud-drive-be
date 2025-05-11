package handler

import (
	"api-gateway/internal/pkg/client"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetUserByID(c *gin.Context) {
	id := c.Query("userId")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": 400,
			"error":      "Bad Request",
			"message":    "Missing userId parameter",
		})
		return
	}

	res, err := client.GetUserId(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": 500,
			"error":      "Failed to get user",
			"message":    err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", res)
}

func GetHealth(c *gin.Context) {
	healthData, err := client.CallUserServiceHealth()

	if err == nil {
		c.Data(http.StatusOK, "application/json", healthData)
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"statusCode": 503,
			"status":     "User service is not responding",
			"service":    "Users services",
		})
	}
}
