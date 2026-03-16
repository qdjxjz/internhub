package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"internhub/pkg/logger"
)

func main() {
	// 初始化日志系统
	if err := logger.Init(); err != nil {
		log.Fatal("failed to init logger")
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	{
		api.POST("/users/register", proxyToAuth("POST", "/api/v1/users/register"))
		api.POST("/users/login", proxyToAuth("POST", "/api/v1/users/login"))

		api.GET("/users/me", JWTMiddleware(), proxyToUserWithAuth("GET", "/api/v1/users/me"))
		api.PATCH("/users/me", JWTMiddleware(), proxyToUserWithAuth("PATCH", "/api/v1/users/me"))

		api.GET("/jobs", proxyToJob())
		api.GET("/jobs/:id", proxyToJob())
		api.POST("/jobs", proxyToJob())

		api.POST("/applications", JWTMiddleware(), proxyToApplyWithAuth("POST", "/api/v1/applications"))
		api.GET("/applications/me", JWTMiddleware(), proxyToApplyWithAuth("GET", "/api/v1/applications/me"))

		api.GET("/recommendations", JWTMiddleware(), proxyToRecommendWithAuth("GET", "/api/v1/recommendations"))

		api.GET("/protected", JWTMiddleware(), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{
				"message": "protected route",
				"user_id": userID,
			})
		})
	}

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

func getAuthServiceURL() string {
	if u := os.Getenv("AUTH_SERVICE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "http://127.0.0.1:8081"
}

func getUserServiceURL() string {
	if u := os.Getenv("USER_SERVICE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "http://127.0.0.1:8082"
}

func getJobServiceURL() string {
	if u := os.Getenv("JOB_SERVICE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "http://127.0.0.1:8083"
}

func getApplyServiceURL() string {
	if u := os.Getenv("APPLY_SERVICE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "http://127.0.0.1:8084"
}

func getRecommendServiceURL() string {
	if u := os.Getenv("RECOMMEND_SERVICE_URL"); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return "http://127.0.0.1:8085"
}

func getJWTSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return []byte("internhub-secret")
	}
	return []byte(s)
}

func proxyToAuth(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("failed to read request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		targetURL := getAuthServiceURL() + path
		req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			logger.Log.Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		req.Header = c.Request.Header.Clone()
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("auth service unavailable: " + err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth service unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

const headerUserID = "X-User-Id"

func proxyToJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		path := "/api/v1/jobs"
		if id != "" {
			path = path + "/" + id
		}
		body, _ := io.ReadAll(c.Request.Body)
		targetURL := getJobServiceURL() + path
		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			logger.Log.Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		req.Header = c.Request.Header.Clone()
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("job service unavailable: " + err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "job service unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

func proxyToApplyWithAuth(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}
		var body io.Reader
		if method != "GET" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			body = bytes.NewBuffer(bodyBytes)
		}
		targetURL := getApplyServiceURL() + path
		if q := c.Request.URL.RawQuery; q != "" {
			targetURL += "?" + q
		}
		req, err := http.NewRequest(method, targetURL, body)
		if err != nil {
			logger.Log.Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		req.Header = c.Request.Header.Clone()
		req.Header.Set(headerUserID, fmt.Sprint(userID))
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("apply service unavailable: " + err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "apply service unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

func proxyToUserWithAuth(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}
		var body io.Reader
		if method != "GET" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			body = bytes.NewBuffer(bodyBytes)
		}
		targetURL := getUserServiceURL() + path
		req, err := http.NewRequest(method, targetURL, body)
		if err != nil {
			logger.Log.Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		req.Header = c.Request.Header.Clone()
		req.Header.Set(headerUserID, fmt.Sprint(userID))
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("user service unavailable: " + err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user service unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

func proxyToRecommendWithAuth(method, path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			return
		}
		var body io.Reader
		if method != "GET" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			body = bytes.NewBuffer(bodyBytes)
		}
		targetURL := getRecommendServiceURL() + path
		if q := c.Request.URL.RawQuery; q != "" {
			targetURL += "?" + q
		}
		req, err := http.NewRequest(method, targetURL, body)
		if err != nil {
			logger.Log.Error("failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		req.Header = c.Request.Header.Clone()
		req.Header.Set(headerUserID, fmt.Sprint(userID))
		// 推荐服务会调 OpenAI，可能较慢，超时设为 60 秒
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Log.Error("recommend service unavailable: " + err.Error())
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recommend service unavailable"})
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

func JWTMiddleware() gin.HandlerFunc {
	secret := getJWTSecret()
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			c.Abort()
			return
		}
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				c.Abort()
				return
			}
		}
		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}