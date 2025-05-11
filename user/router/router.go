package router

import (
	"github.com/gin-gonic/gin"
	"user-service/internal/user/handler"
	"user-service/internal/user/repository"
	"user-service/internal/user/service"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// DI
	repo := repository.NewInMemoryUserRepo()
	svc := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "user-service healthy"})
	})

	r.GET("/users/GetById?userId=", userHandler.GetUserByID)

	return r
}
