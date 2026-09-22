package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// OptionInput 选项输入。
type OptionInput struct {
	Key  string `json:"key" binding:"required,min=1,max=2"`
	Text string `json:"text" binding:"required,min=1,max=500"`
}

// CreateQuestionRequest 创建题目请求（题型枚举校验：single/multiple/judge/fill/short）。
type CreateQuestionRequest struct {
	Type            string        `json:"type" binding:"required,oneof=single multiple judge fill short"`
	Subject         string        `json:"subject" binding:"required,min=1,max=100"`
	KnowledgePoints []string      `json:"knowledge_points" binding:"required,min=1,max=10"`
	Difficulty      string        `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Content         string        `json:"content" binding:"required,min=1,max=2000"`
	Options         []OptionInput `json:"options"`
	Answer          string        `json:"answer" binding:"required,max=500"`
	Analysis        string        `json:"analysis" binding:"omitempty,max=2000"`
	Score           float64       `json:"score" binding:"required,min=0.5,max=100"`
	Status          string        `json:"status" binding:"omitempty,oneof=draft published"`
}

// UpdateQuestionRequest 更新题目请求。
type UpdateQuestionRequest struct {
	Type            string        `json:"type" binding:"omitempty,oneof=single multiple judge fill short"`
	Subject         string        `json:"subject" binding:"omitempty,min=1,max=100"`
	KnowledgePoints []string      `json:"knowledge_points" binding:"omitempty,max=10"`
	Difficulty      string        `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	Content         string        `json:"content" binding:"omitempty,min=1,max=2000"`
	Options         []OptionInput `json:"options"`
	Answer          string        `json:"answer" binding:"omitempty,max=500"`
	Analysis        string        `json:"analysis" binding:"omitempty,max=2000"`
	Score           float64       `json:"score" binding:"omitempty,min=0.5,max=100"`
	Status          string        `json:"status" binding:"omitempty,oneof=draft published"`
}

// QuestionQuery 题目查询参数。
type QuestionQuery struct {
	Subject         string   `form:"subject"`
	Type            string   `form:"type"`
	Difficulty      string   `form:"difficulty"`
	KnowledgePoint  string   `form:"knowledge_point"`
	Keyword         string   `form:"keyword"`
	Status          string   `form:"status"`
	Page            int64    `form:"page"`
	PageSize        int64    `form:"page_size"`
}

// QuestionResponse 题目响应。
type QuestionResponse struct {
	ID              string             `json:"id"`
	Type            string             `json:"type"`
	Subject         string             `json:"subject"`
	KnowledgePoints []string           `json:"knowledge_points"`
	Difficulty      string             `json:"difficulty"`
	Content         string             `json:"content"`
	Options         []model.QuestionOption `json:"options"`
	Answer          string             `json:"answer"`
	Analysis        string             `json:"analysis"`
	Score           float64            `json:"score"`
	Status          string             `json:"status"`
	CreatorID       string             `json:"creator_id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// ToQuestionResponse 模型转响应。
func ToQuestionResponse(q *model.Question) QuestionResponse {
	return QuestionResponse{
		ID:              q.ID.Hex(),
		Type:            q.Type,
		Subject:         q.Subject,
		KnowledgePoints: q.KnowledgePoints,
		Difficulty:      q.Difficulty,
		Content:         q.Content,
		Options:         q.Options,
		Answer:          q.Answer,
		Analysis:        q.Analysis,
		Score:           q.Score,
		Status:          q.Status,
		CreatorID:       q.CreatorID.Hex(),
		CreatedAt:       q.CreatedAt,
		UpdatedAt:       q.UpdatedAt,
	}
}
