package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"internhub/user-service/internal/service"
)

const HeaderUserID = "X-User-Id"

func GetMe(c *gin.Context) {
	userIDStr := c.GetHeader(HeaderUserID)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	p, err := service.GetProfile(uint(userID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"profile": nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": p.UserID,
		"profile": p,
	})
}

type UpdateMeRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func UpdateMe(c *gin.Context) {
	userIDStr := c.GetHeader(HeaderUserID)
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := service.UpdateProfile(uint(userID), req.Nickname, req.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": p})
}
