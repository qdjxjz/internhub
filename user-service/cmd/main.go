package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"internhub/user-service/config"
	"internhub/user-service/internal/handler"
	"internhub/user-service/internal/model"
)

func main() {
	config.InitDB()

	if err := config.DB.AutoMigrate(&model.UserProfile{}); err != nil {
		log.Fatalf("AutoMigrate fail: %v", err)
	}

	r := gin.Default()
	api := r.Group("/api/v1")
	{
		api.GET("/users/me", handler.GetMe)
		api.PATCH("/users/me", handler.UpdateMe)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	r.Run(":" + port)
}
