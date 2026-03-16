package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"internhub/recommend-service/internal/client"
	"internhub/recommend-service/internal/service"
)

func TestGetRecommendations_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	// 不设置 X-User-Id

	mock := &mockRecommender{}
	GetRecommendations(mock)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if mock.called {
		t.Error("GetRecommendations should not be called when user id missing")
	}
}

func TestGetRecommendations_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	c.Request.Header.Set(HeaderUserID, "abc")

	mock := &mockRecommender{}
	GetRecommendations(mock)(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if mock.called {
		t.Error("GetRecommendations should not be called when user id invalid")
	}
}

func TestGetRecommendations_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	c.Request.Header.Set(HeaderUserID, "1")

	mock := &mockRecommender{err: http.ErrHandlerTimeout}
	GetRecommendations(mock)(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if !mock.called || mock.lastUserID != 1 {
		t.Errorf("mock: called=%v lastUserID=%d", mock.called, mock.lastUserID)
	}
}

func TestGetRecommendations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	c.Request.Header.Set(HeaderUserID, "42")

	result := &service.RecommendResult{
		List: []service.RecommendedItem{
			{Job: client.Job{ID: 1, Title: "Go 实习", Company: "A", Link: "https://a.com"}},
		},
		Summary: "推荐",
	}
	mock := &mockRecommender{result: result}
	GetRecommendations(mock)(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !mock.called || mock.lastUserID != 42 {
		t.Errorf("mock: called=%v lastUserID=%d", mock.called, mock.lastUserID)
	}
	// 简单校验 body 包含职位信息
	body := w.Body.String()
	if body == "" || len(body) < 10 {
		t.Errorf("unexpected body: %s", body)
	}
}

type mockRecommender struct {
	called     bool
	lastUserID uint
	result     *service.RecommendResult
	err        error
}

func (m *mockRecommender) GetRecommendations(userID uint) (*service.RecommendResult, error) {
	m.called = true
	m.lastUserID = userID
	return m.result, m.err
}
