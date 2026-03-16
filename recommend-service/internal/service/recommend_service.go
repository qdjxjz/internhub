package service

import (
	"fmt"

	"go.uber.org/zap"
	"internhub/pkg/logger"
	"internhub/recommend-service/config"
	"internhub/recommend-service/internal/ai"
	"internhub/recommend-service/internal/client"
)

// RecommendedItem 返回给前端的单条推荐（职位 + 可选理由）
type RecommendedItem struct {
	Job    client.Job `json:"job"`
	Reason string     `json:"reason,omitempty"`
}

// RecommendResult 推荐接口返回结构
type RecommendResult struct {
	List    []RecommendedItem `json:"list"`
	Summary string            `json:"summary,omitempty"`
	AiUsed  bool              `json:"ai_used"` // 本次是否使用了 AI（未配置或调用失败时为 false）
}

// GetRecommendations 获取当前用户的岗位推荐（已排除已投递，可选 AI 排序与理由）
func GetRecommendations(userID uint) (*RecommendResult, error) {
	jobs, err := client.FetchJobs()
	if err != nil {
		return nil, fmt.Errorf("fetch jobs: %w", err)
	}
	applications, err := client.FetchUserApplications(userID)
	if err != nil {
		return nil, fmt.Errorf("fetch applications: %w", err)
	}
	appliedSet := make(map[uint]bool)
	for _, a := range applications {
		appliedSet[a.JobID] = true
	}
	jobMap := make(map[uint]client.Job)
	var candidates []client.Job
	for _, j := range jobs {
		jobMap[j.ID] = j
		if !appliedSet[j.ID] {
			candidates = append(candidates, j)
		}
	}
	if len(candidates) == 0 {
		return &RecommendResult{List: []RecommendedItem{}, AiUsed: false}, nil
	}

	appliedIDs := make([]uint, 0, len(applications))
	for _, a := range applications {
		appliedIDs = append(appliedIDs, a.JobID)
	}
	nickname := ""
	if profile, _ := client.FetchUserProfile(userID); profile != nil {
		nickname = profile.Nickname
	}

	if config.RecommendEnabled && config.OpenAIAPIKey != "" {
		rec, err := ai.Recommend(config.OpenAIAPIKey, config.OpenAIBaseURL, ai.RecommendRequest{
			Jobs:         candidates,
			AppliedIDs:   appliedIDs,
			UserNickname: nickname,
		})
		if err != nil {
			logger.Log.Error("ai recommend failed, fallback to order by time", zap.String("error", err.Error()))
			// 降级：按创建时间倒序返回未投递列表
			return buildFallbackResultWithHint(candidates, true), nil
		}
		out, err := buildResultFromAI(rec, jobMap, candidates)
		if err != nil {
			return buildFallbackResultWithHint(candidates, true), nil
		}
		out.AiUsed = true
		return out, nil
	}
	return buildFallbackResultWithHint(candidates, false), nil
}

func buildFallbackResultWithHint(candidates []client.Job, aiFailed bool) *RecommendResult {
	list := make([]RecommendedItem, 0, len(candidates))
	for _, j := range candidates {
		list = append(list, RecommendedItem{Job: j})
	}
	hint := ""
	if aiFailed {
		hint = "（AI 暂时不可用，已按时间排序）"
	}
	return &RecommendResult{List: list, Summary: hint, AiUsed: false}
}

func buildResultFromAI(rec *ai.RecommendResponse, jobMap map[uint]client.Job, candidates []client.Job) (*RecommendResult, error) {
	list := make([]RecommendedItem, 0, len(rec.JobIDs))
	seen := make(map[uint]bool)
	for _, id := range rec.JobIDs {
		if seen[id] {
			continue
		}
		j, ok := jobMap[id]
		if !ok {
			continue
		}
		seen[id] = true
		item := RecommendedItem{Job: j}
		if rec.Reasons != nil {
			item.Reason = rec.Reasons[id]
		}
		list = append(list, item)
	}
	// AI 可能只返回了部分 ID，把未出现在 list 里的候选也追加到末尾
	for _, j := range candidates {
		if !seen[j.ID] {
			list = append(list, RecommendedItem{Job: j})
		}
	}
	return &RecommendResult{
		List:    list,
		Summary: rec.Summary,
	}, nil
}
