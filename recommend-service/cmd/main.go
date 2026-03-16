package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"internhub/pkg/logger"
	"internhub/recommend-service/config"
	"internhub/recommend-service/internal/handler"
	"internhub/recommend-service/internal/service"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatal("failed to init logger")
	}
	config.Init()
	if config.RecommendEnabled {
		logger.Log.Info("AI 推荐已启用（OPENAI_API_KEY 已配置）")
	} else {
		logger.Log.Info("AI 推荐未启用（未配置 OPENAI_API_KEY 或未生效）；推荐接口将按时间排序返回")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.GET("/recommendations", handler.GetRecommendations(handler.RecommenderFunc(service.GetRecommendations)))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}
	logger.Log.Info("recommend-service listening", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
