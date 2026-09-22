package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// ExamQuestionInput 组卷题目输入。
type ExamQuestionInput struct {
	QuestionID string  `json:"question_id" binding:"required"`
	Score      float64 `json:"score" binding:"omitempty,min=0.5,max=100"`
}

// CreateExamRequest 手动创建试卷请求。
type CreateExamRequest struct {
	Title           string              `json:"title" binding:"required,min=2,max=200"`
	Subject         string              `json:"subject" binding:"required,min=1,max=100"`
	Description     string              `json:"description" binding:"omitempty,max=2000"`
	TotalScore      float64             `json:"total_score" binding:"omitempty,min=1,max=1000"`
	PassScore       float64             `json:"pass_score" binding:"omitempty,min=0,max=1000"`
	DurationMin     int                 `json:"duration_min" binding:"required,min=1,max=600"`
	StartAt         time.Time           `json:"start_at" binding:"required"`
	EndAt           time.Time           `json:"end_at" binding:"required"`
	ShuffleQuestion bool                `json:"shuffle_question"`
	ShuffleOption   bool                `json:"shuffle_option"`
	Questions       []ExamQuestionInput `json:"questions"`
}

// UpdateExamRequest 更新试卷请求（状态字段校验：draft/published）。
type UpdateExamRequest struct {
	Title           string              `json:"title" binding:"omitempty,min=2,max=200"`
	Subject         string              `json:"subject" binding:"omitempty,min=1,max=100"`
	Description     string              `json:"description" binding:"omitempty,max=2000"`
	TotalScore      float64             `json:"total_score" binding:"omitempty,min=1,max=1000"`
	PassScore       float64             `json:"pass_score" binding:"omitempty,min=0,max=1000"`
	DurationMin     int                 `json:"duration_min" binding:"omitempty,min=1,max=600"`
	StartAt         *time.Time          `json:"start_at"`
	EndAt           *time.Time          `json:"end_at"`
	ShuffleQuestion *bool               `json:"shuffle_question"`
	ShuffleOption   *bool               `json:"shuffle_option"`
	Questions       []ExamQuestionInput `json:"questions"`
	Status          string              `json:"status" binding:"omitempty,oneof=draft published"`
}

// AutoGenerateRequest 自动组卷请求。
type AutoGenerateRequest struct {
	Title           string    `json:"title" binding:"required,min=2,max=200"`
	Subject         string    `json:"subject" binding:"required,min=1,max=100"`
	Description     string    `json:"description" binding:"omitempty,max=2000"`
	TotalScore      float64   `json:"total_score" binding:"omitempty,min=1,max=1000"`
	PassScore       float64   `json:"pass_score" binding:"omitempty,min=0,max=1000"`
	DurationMin     int       `json:"duration_min" binding:"required,min=1,max=600"`
	StartAt         time.Time `json:"start_at" binding:"required"`
	EndAt           time.Time `json:"end_at" binding:"required"`
	ShuffleQuestion bool      `json:"shuffle_question"`
	ShuffleOption   bool      `json:"shuffle_option"`
	// 组卷策略：知识点覆盖与难度分布
	KnowledgePoints  []string       `json:"knowledge_points" binding:"required,min=1,max=20"`
	DifficultyDist   map[string]int `json:"difficulty_dist" binding:"required"` // 如 {"easy":5,"medium":5,"hard":3}
	ScorePerQuestion float64        `json:"score_per_question" binding:"required,min=0.5,max=50"`
}

// ExamQuery 试卷查询参数。
type ExamQuery struct {
	Title    string `form:"title"`
	Subject  string `form:"subject"`
	Status   string `form:"status"`
	Page     int64  `form:"page"`
	PageSize int64  `form:"page_size"`
}

// ExamQuestionResponse 试卷题目响应。
type ExamQuestionResponse struct {
	QuestionID string  `json:"question_id"`
	Score      float64 `json:"score"`
	Order      int     `json:"order"`
}

// ExamResponse 试卷响应。
type ExamResponse struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Subject         string                 `json:"subject"`
	Description     string                 `json:"description"`
	TotalScore      float64                `json:"total_score"`
	PassScore       float64                `json:"pass_score"`
	DurationMin     int                    `json:"duration_min"`
	StartAt         time.Time              `json:"start_at"`
	EndAt           time.Time              `json:"end_at"`
	Status          string                 `json:"status"`
	ShuffleQuestion bool                   `json:"shuffle_question"`
	ShuffleOption   bool                   `json:"shuffle_option"`
	Questions       []ExamQuestionResponse `json:"questions"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ToExamResponse 模型转响应。
func ToExamResponse(e *model.Exam) ExamResponse {
	resp := ExamResponse{
		ID:              e.ID.Hex(),
		Title:           e.Title,
		Subject:         e.Subject,
		Description:     e.Description,
		TotalScore:      e.TotalScore,
		PassScore:       e.PassScore,
		DurationMin:     e.DurationMin,
		StartAt:         e.StartAt,
		EndAt:           e.EndAt,
		Status:          e.Status,
		ShuffleQuestion: e.ShuffleQuestion,
		ShuffleOption:   e.ShuffleOption,
		CreatedBy:       e.CreatedBy.Hex(),
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
	for _, q := range e.Questions {
		resp.Questions = append(resp.Questions, ExamQuestionResponse{
			QuestionID: q.QuestionID.Hex(),
			Score:      q.Score,
			Order:      q.Order,
		})
	}
	return resp
}
