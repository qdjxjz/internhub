package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"internhub/pkg/logger"
	"internhub/recommend-service/config"
)

// Job 与 job-service 的 model 一致，用于反序列化
type Job struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	Link      string `json:"link"`
	CreatedAt string `json:"created_at"`
}

type JobsResponse struct {
	List []Job `json:"list"`
}

// Application 与 apply-service 的 model 一致
type Application struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	JobID  uint   `json:"job_id"`
	Status string `json:"status"`
}

type ApplicationsResponse struct {
	List []Application `json:"list"`
}

// UserProfile 与 user-service 的 profile 一致
type UserProfile struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// FetchJobs 从 job-service 拉取职位列表
func FetchJobs() ([]Job, error) {
	url := config.JobServiceURL + "/api/v1/jobs"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Log.Error("recommend: job service request failed", zapError(err))
		return nil, fmt.Errorf("job service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("job service returned %d: %s", resp.StatusCode, string(body))
	}
	var out JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.List, nil
}

// FetchUserApplications 从 apply-service 拉取用户投递列表（需带 X-User-Id）
func FetchUserApplications(userID uint) ([]Application, error) {
	url := config.ApplyServiceURL + "/api/v1/applications/me"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Id", fmt.Sprintf("%d", userID))
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Log.Error("recommend: apply service request failed", zapError(err))
		return nil, fmt.Errorf("apply service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("apply service returned %d: %s", resp.StatusCode, string(body))
	}
	var out ApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.List, nil
}

// FetchUserProfile 从 user-service 拉取用户资料（可选，用于 AI 上下文）
func FetchUserProfile(userID uint) (*UserProfile, error) {
	url := config.UserServiceURL + "/api/v1/users/me"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Id", fmt.Sprintf("%d", userID))
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Log.Warn("recommend: user service request failed (optional)", zapError(err))
		return nil, nil // 非致命
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	var out UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, nil
	}
	return &out, nil
}

func zapError(e error) zap.Field {
	return zap.String("error", e.Error())
}
