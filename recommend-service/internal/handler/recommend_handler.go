package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"internhub/recommend-service/internal/service"
)

const HeaderUserID = "X-User-Id"

// Recommender 推荐逻辑抽象，便于测试注入 mock
type Recommender interface {
	GetRecommendations(userID uint) (*service.RecommendResult, error)
}

// RecommenderFunc 使函数满足 Recommender 接口
type RecommenderFunc func(userID uint) (*service.RecommendResult, error)

func (f RecommenderFunc) GetRecommendations(userID uint) (*service.RecommendResult, error) {
	return f(userID)
}

// GetRecommendations 返回需要注入 Recommender 的 Gin 处理函数
func GetRecommendations(rec Recommender) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := c.GetHeader(HeaderUserID)
		if s == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}
		userID, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		result, err := rec.GetRecommendations(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
