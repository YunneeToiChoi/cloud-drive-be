package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"user-service/internal/user/service"
)

type UserHandler struct {
	svc service.IUserService
}

func NewUserHandler(svc service.IUserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	var id string

	id = c.Param("id")

	if id == "" {
		id = c.Query("userId")
	}

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": 400,
			"error":      "User ID not provided",
		})
		return
	}

	user, err := h.svc.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"statusCode": 404,
			"error":      "User not found",
		})
		return
	}

	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"data":       user,
	})
}
