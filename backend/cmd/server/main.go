package main

import (
	"os"

	"cyskillswap/internal/logger"
	"cyskillswap/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "19629"
	}
	r := gin.Default()
	routes.Register(r)
	logger.Info("server started", port)
	r.Run(":" + port)
}
