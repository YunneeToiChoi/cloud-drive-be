package cmd

import (
	"user-service/router"
)

func main() {
	r := router.SetupRouter()
	err := r.Run(":8081")
	if err != nil {
		return
	}
}
