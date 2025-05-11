package cmd

import (
	"api-gateway/internal/router"
	"log"
)

func main() {
	r := router.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
	
}
