package config

import (
	"os"
	"strings"
)

var (
	JobServiceURL    string
	ApplyServiceURL  string
	UserServiceURL   string
	OpenAIAPIKey     string
	OpenAIBaseURL    string // 可选，兼容 OpenAI 兼容接口（如 Azure、国内代理）
	RecommendEnabled bool   // 是否启用 AI 推荐（未配置 API Key 时仅返回未投递列表）
)

func Init() {
	JobServiceURL = getURL("JOB_SERVICE_URL", "http://127.0.0.1:8083")
	ApplyServiceURL = getURL("APPLY_SERVICE_URL", "http://127.0.0.1:8084")
	UserServiceURL = getURL("USER_SERVICE_URL", "http://127.0.0.1:8082")
	OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	OpenAIBaseURL = os.Getenv("OPENAI_BASE_URL") // 空则使用 OpenAI 默认
	RecommendEnabled = OpenAIAPIKey != ""
}

func getURL(key, defaultVal string) string {
	if u := os.Getenv(key); u != "" {
		return strings.TrimSuffix(u, "/")
	}
	return defaultVal
}
