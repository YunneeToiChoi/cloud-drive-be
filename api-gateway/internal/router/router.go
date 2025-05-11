package router

import (
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())

	// Health check endpoints
	r.GET("/health", handler.HealthCheck)
	r.GET("/health/services", handler.HealthCheckServices)

	api := r.Group("/api")
	{
		// User endpoints
		users := api.Group("/users")
		{
			users.GET("/health", handler.GetHealth)
			users.GET("/GetById", handler.GetUserByID)
		}
	}

	return r
}
