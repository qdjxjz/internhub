package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"internhub/recommend-service/config"
	"internhub/recommend-service/internal/ai"
	"internhub/recommend-service/internal/client"
)

func TestGetRecommendations_ExcludesAppliedJobs(t *testing.T) {
	// 避免测试时调用真实 OpenAI
	os.Setenv("OPENAI_API_KEY", "")
	defer os.Unsetenv("OPENAI_API_KEY")

	jobsResp := map[string]interface{}{
		"list": []client.Job{
			{ID: 1, Title: "Job1", Company: "C1", Link: ""},
			{ID: 2, Title: "Job2", Company: "C2", Link: ""},
			{ID: 3, Title: "Job3", Company: "C3", Link: ""},
		},
	}
	jobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobsResp)
	}))
	defer jobSrv.Close()

	// 用户已投递 1 和 3
	applyResp := map[string]interface{}{
		"list": []client.Application{
			{ID: 1, UserID: 100, JobID: 1, Status: "pending"},
			{ID: 2, UserID: 100, JobID: 3, Status: "pending"},
		},
	}
	applySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-Id") != "100" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(applyResp)
	}))
	defer applySrv.Close()

	userSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"user_id": 100, "nickname": "Test"})
	}))
	defer userSrv.Close()

	os.Setenv("JOB_SERVICE_URL", jobSrv.URL)
	os.Setenv("APPLY_SERVICE_URL", applySrv.URL)
	os.Setenv("USER_SERVICE_URL", userSrv.URL)
	config.Init()

	result, err := GetRecommendations(100)
	if err != nil {
		t.Fatalf("GetRecommendations: %v", err)
	}
	// 应只包含未投递的 Job2
	if len(result.List) != 1 {
		t.Fatalf("expected 1 recommended job, got %d: %+v", len(result.List), result.List)
	}
	if result.List[0].Job.ID != 2 {
		t.Errorf("expected job id 2, got %d", result.List[0].Job.ID)
	}
}

func TestGetRecommendations_EmptyJobs(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "")
	defer os.Unsetenv("OPENAI_API_KEY")

	jobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []client.Job{}})
	}))
	defer jobSrv.Close()
	applySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"list": []client.Application{}})
	}))
	defer applySrv.Close()

	os.Setenv("JOB_SERVICE_URL", jobSrv.URL)
	os.Setenv("APPLY_SERVICE_URL", applySrv.URL)
	config.Init()

	result, err := GetRecommendations(1)
	if err != nil {
		t.Fatalf("GetRecommendations: %v", err)
	}
	if len(result.List) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(result.List))
	}
}

func TestGetRecommendations_AllApplied(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "")
	defer os.Unsetenv("OPENAI_API_KEY")

	jobSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"list": []client.Job{{ID: 1, Title: "T", Company: "C", Link: ""}},
		})
	}))
	defer jobSrv.Close()
	applySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"list": []client.Application{{ID: 1, UserID: 1, JobID: 1, Status: "pending"}},
		})
	}))
	defer applySrv.Close()

	os.Setenv("JOB_SERVICE_URL", jobSrv.URL)
	os.Setenv("APPLY_SERVICE_URL", applySrv.URL)
	config.Init()

	result, err := GetRecommendations(1)
	if err != nil {
		t.Fatalf("GetRecommendations: %v", err)
	}
	if len(result.List) != 0 {
		t.Errorf("expected 0 when all applied, got %d", len(result.List))
	}
}

func TestBuildFallbackResult(t *testing.T) {
	candidates := []client.Job{
		{ID: 1, Title: "A", Company: "C1", Link: ""},
		{ID: 2, Title: "B", Company: "C2", Link: ""},
	}
	out := buildFallbackResultWithHint(candidates, false)
	if len(out.List) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.List))
	}
	if out.List[0].Job.ID != 1 || out.List[1].Job.ID != 2 {
		t.Errorf("unexpected order: %+v", out.List)
	}
	if out.Summary != "" {
		t.Errorf("fallback(aiFailed=false) should have empty summary, got %q", out.Summary)
	}
	if out.AiUsed {
		t.Error("fallback should have AiUsed false")
	}
}

func TestBuildResultFromAI(t *testing.T) {
	jobMap := map[uint]client.Job{
		1: {ID: 1, Title: "A", Company: "C1", Link: ""},
		2: {ID: 2, Title: "B", Company: "C2", Link: ""},
		3: {ID: 3, Title: "C", Company: "C3", Link: ""},
	}
	candidates := []client.Job{jobMap[1], jobMap[2], jobMap[3]}

	rec := &ai.RecommendResponse{
		JobIDs:  []uint{2, 1},
		Reasons: map[uint]string{1: "r1", 2: "r2"},
		Summary: "summary",
	}
	out, err := buildResultFromAI(rec, jobMap, candidates)
	if err != nil {
		t.Fatalf("buildResultFromAI: %v", err)
	}
	if len(out.List) != 3 {
		t.Fatalf("expected 3 items (2 from AI + 1 remaining), got %d", len(out.List))
	}
	if out.List[0].Job.ID != 2 || out.List[0].Reason != "r2" {
		t.Errorf("first item: id=%d reason=%q", out.List[0].Job.ID, out.List[0].Reason)
	}
	if out.List[1].Job.ID != 1 || out.List[1].Reason != "r1" {
		t.Errorf("second item: id=%d reason=%q", out.List[1].Job.ID, out.List[1].Reason)
	}
	if out.List[2].Job.ID != 3 {
		t.Errorf("third should be remaining candidate 3, got %d", out.List[2].Job.ID)
	}
	if out.Summary != "summary" {
		t.Errorf("summary: got %q", out.Summary)
	}
}
