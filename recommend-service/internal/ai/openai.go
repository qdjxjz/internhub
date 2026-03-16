package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"internhub/pkg/logger"
	"internhub/recommend-service/internal/client"
)

const defaultOpenAIBase = "https://api.openai.com/v1"

// RecommendRequest 用于构建发给 AI 的上下文
type RecommendRequest struct {
	Jobs         []client.Job
	AppliedIDs   []uint
	UserNickname string
}

// Recommendation 单条推荐结果（含 AI 理由）
type Recommendation struct {
	JobID  uint   `json:"job_id"`
	Reason string `json:"reason,omitempty"`
}

// RecommendResponse AI 返回的推荐结果（解析自 content）
type RecommendResponse struct {
	JobIDs  []uint            `json:"job_ids"`  // 推荐顺序的职位 ID 列表
	Reasons map[uint]string   `json:"reasons"`  // job_id -> 推荐理由（可选）
	Summary string            `json:"summary"`  // 总体说明（可选）
}

// openaiMessage 与 API 格式一致
type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiReq struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Recommend 调用 OpenAI 兼容接口，根据职位列表与用户上下文返回推荐顺序与理由
func Recommend(apiKey, baseURL string, req RecommendRequest) (*RecommendResponse, error) {
	if apiKey == "" {
		return nil, nil
	}
	if baseURL == "" {
		baseURL = defaultOpenAIBase
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	systemContent := `你是一个实习/校招岗位推荐助手。用户会收到一段 JSON，包含可选职位列表、用户已投递的职位 ID、以及用户昵称（可能为空）。
你需要根据职位标题、公司名称，结合“未投递”的约束，给出推荐顺序和简短理由。
请仅用 JSON 回复，不要其他说明。格式如下（不要省略字段）：
{"job_ids": [推荐的职位ID按推荐顺序], "reasons": {"职位ID": "一句话理由"}, "summary": "整体推荐说明一句话"}`

	jobList := make([]map[string]interface{}, 0, len(req.Jobs))
	for _, j := range req.Jobs {
		jobList = append(jobList, map[string]interface{}{
			"id":      j.ID,
			"title":   j.Title,
			"company": j.Company,
			"link":    j.Link,
		})
	}
	applied := req.AppliedIDs
	if applied == nil {
		applied = []uint{}
	}
	userJSON := map[string]interface{}{
		"jobs":          jobList,
		"applied_job_ids": applied,
		"user_nickname": req.UserNickname,
	}
	userBytes, _ := json.Marshal(userJSON)
	userContent := string(userBytes)

	body := openaiReq{
		Model: "gpt-4o-mini",
		Messages: []openaiMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	hc := &http.Client{Timeout: 30 * time.Second}
	resp, err := hc.Do(httpReq)
	if err != nil {
		logger.Log.Error("openai request failed", zap.String("error", err.Error()))
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("openai non-200", zap.Int("status", resp.StatusCode), zap.ByteString("body", respBody))
		return nil, fmt.Errorf("openai returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out openaiResp
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("openai empty choices")
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	// 去除可能的 markdown 代码块
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	// 模型有时在末尾多一个反引号，导致 JSON 解析失败
	content = strings.TrimRight(content, "`")
	content = strings.TrimSpace(content)
	var rec RecommendResponse
	if err := json.Unmarshal([]byte(content), &rec); err != nil {
		logger.Log.Error("openai parse response failed", zap.String("content", content))
		return nil, err
	}
	return &rec, nil
}
