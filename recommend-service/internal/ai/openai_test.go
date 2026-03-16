package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"internhub/pkg/logger"
	"internhub/recommend-service/internal/client"
)

func TestMain(m *testing.M) {
	_ = logger.Init()
	os.Exit(m.Run())
}

func TestRecommend_EmptyAPIKey(t *testing.T) {
	got, err := Recommend("", "", RecommendRequest{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result when apiKey empty, got %+v", got)
	}
}

func TestRecommend_ValidResponse(t *testing.T) {
	responseBody := `{"choices":[{"message":{"content":"{\"job_ids\":[2,1],\"reasons\":{\"1\":\"适合后端\",\"2\":\"大厂实习\"},\"summary\":\"推荐以上两则\"}"}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	baseURL := server.URL
	req := RecommendRequest{
		Jobs: []client.Job{
			{ID: 1, Title: "Go 实习", Company: "A 公司", Link: "https://a.com"},
			{ID: 2, Title: "后端实习", Company: "B 公司", Link: "https://b.com"},
		},
		AppliedIDs:   []uint{},
		UserNickname: "Test",
	}
	rec, err := Recommend("sk-fake", baseURL, req)
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if rec == nil {
		t.Fatal("expected non-nil result")
	}
	if len(rec.JobIDs) != 2 {
		t.Errorf("expected 2 job_ids, got %v", rec.JobIDs)
	}
	if rec.JobIDs[0] != 2 || rec.JobIDs[1] != 1 {
		t.Errorf("expected job_ids [2,1], got %v", rec.JobIDs)
	}
	if rec.Reasons == nil {
		t.Fatal("expected non-nil reasons")
	}
	if rec.Reasons[1] != "适合后端" || rec.Reasons[2] != "大厂实习" {
		t.Errorf("unexpected reasons: %v", rec.Reasons)
	}
	if rec.Summary != "推荐以上两则" {
		t.Errorf("summary: got %q", rec.Summary)
	}
}

func TestRecommend_ResponseWithMarkdownBlock(t *testing.T) {
	content := `{"job_ids":[1],"reasons":{"1":"好"},"summary":"ok"}`
	// 构造 AI 返回的 ```json\n...\n``` 形式，JSON 内用 \n 转义
	escapedContent := "```json\n" + content + "\n```"
	inner, _ := json.Marshal(escapedContent)
	responseBody := `{"choices":[{"message":{"content":` + string(inner) + `}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	rec, err := Recommend("sk-fake", server.URL, RecommendRequest{
		Jobs: []client.Job{{ID: 1, Title: "T", Company: "C", Link: ""}},
	})
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if rec == nil || len(rec.JobIDs) != 1 || rec.JobIDs[0] != 1 {
		t.Errorf("expected job_ids [1], got %+v", rec)
	}
	if rec.Summary != "ok" {
		t.Errorf("summary: got %q", rec.Summary)
	}
}

func TestRecommend_EmptyChoices(t *testing.T) {
	responseBody := `{"choices":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	rec, err := Recommend("sk-fake", server.URL, RecommendRequest{Jobs: []client.Job{}})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if rec != nil {
		t.Errorf("expected nil result, got %+v", rec)
	}
}

func TestRecommend_InvalidJSONInContent(t *testing.T) {
	// content 需为合法 JSON 字符串（带引号），否则外层 JSON 解析会失败
	responseBody := `{"choices":[{"message":{"content":"not json at all"}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	rec, err := Recommend("sk-fake", server.URL, RecommendRequest{Jobs: []client.Job{}})
	if err == nil {
		t.Fatal("expected error for invalid JSON content")
	}
	if rec != nil {
		t.Errorf("expected nil result, got %+v", rec)
	}
}

func TestRecommend_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	rec, err := Recommend("sk-fake", server.URL, RecommendRequest{})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if rec != nil {
		t.Errorf("expected nil result, got %+v", rec)
	}
}

// 确保 RecommendResponse 的 JSON 与 AI 返回格式一致（reasons 为 map[uint]string 时需数字 key）
func TestRecommendResponse_JSONUnmarshal(t *testing.T) {
	// OpenAI 返回的 JSON 里 key 是字符串 "1", "2"
	raw := `{"job_ids":[2,1],"reasons":{"1":"a","2":"b"},"summary":"s"}`
	var rec RecommendResponse
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(rec.JobIDs) != 2 || rec.JobIDs[0] != 2 || rec.JobIDs[1] != 1 {
		t.Errorf("job_ids: %v", rec.JobIDs)
	}
	// Go 的 json 会把 "1","2" 解析到 map[uint]string 的 1, 2
	if rec.Reasons[1] != "a" || rec.Reasons[2] != "b" {
		t.Errorf("reasons: %v", rec.Reasons)
	}
	if rec.Summary != "s" {
		t.Errorf("summary: %q", rec.Summary)
	}
}
