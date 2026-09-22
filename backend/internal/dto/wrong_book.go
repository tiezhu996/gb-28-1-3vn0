package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// AddWrongBookRequest 添加错题请求。
type AddWrongBookRequest struct {
	QuestionID   string `json:"question_id" binding:"required"`
	ExamID       string `json:"exam_id" binding:"omitempty"`
	ExamRecordID string `json:"exam_record_id" binding:"omitempty"`
	Note         string `json:"note" binding:"omitempty,max=500"`
}

// UpdateWrongBookRequest 更新错题本（状态字段校验：active/resolved）。
type UpdateWrongBookRequest struct {
	Status string `json:"status" binding:"omitempty,oneof=active resolved"`
	Note   string `json:"note" binding:"omitempty,max=500"`
}

// WrongBookQuery 错题本查询参数。
type WrongBookQuery struct {
	Subject        string `form:"subject"`
	KnowledgePoint string `form:"knowledge_point"`
	Status         string `form:"status"`
	Page           int64  `form:"page"`
	PageSize       int64  `form:"page_size"`
}

// WrongBookResponse 错题本响应。
type WrongBookResponse struct {
	ID              string    `json:"id"`
	StudentID       string    `json:"student_id"`
	QuestionID      string    `json:"question_id"`
	ExamID          string    `json:"exam_id"`
	ExamRecordID    string    `json:"exam_record_id"`
	Subject         string    `json:"subject"`
	KnowledgePoints []string  `json:"knowledge_points"`
	QuestionContent string    `json:"question_content"`
	MyAnswer        string    `json:"my_answer"`
	CorrectAnswer   string    `json:"correct_answer"`
	Analysis        string    `json:"analysis"`
	Note            string    `json:"note"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToWrongBookResponse 模型转响应。
func ToWrongBookResponse(w *model.WrongBook) WrongBookResponse {
	return WrongBookResponse{
		ID:              w.ID.Hex(),
		StudentID:       w.StudentID.Hex(),
		QuestionID:      w.QuestionID.Hex(),
		ExamID:          w.ExamID.Hex(),
		ExamRecordID:    w.ExamRecordID.Hex(),
		Subject:         w.Subject,
		KnowledgePoints: w.KnowledgePoints,
		QuestionContent: w.QuestionContent,
		MyAnswer:        w.MyAnswer,
		CorrectAnswer:   w.CorrectAnswer,
		Analysis:        w.Analysis,
		Note:            w.Note,
		Status:          w.Status,
		CreatedAt:       w.CreatedAt,
	}
}
