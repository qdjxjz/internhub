package main

import (
	"log"
	"net/http"

	"internhub/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 初始化日志系统
	if err := logger.Init(); err != nil {
		log.Fatal("failed to init logger")
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		logger.Log.Info("health check called")
		c.JSON(http.StatusOK, gin.H{
			"status": "gateway ok",
		})
	})
	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.Run(":8080")
}