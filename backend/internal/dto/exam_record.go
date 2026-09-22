package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// StartExamRequest 开始考试请求（无业务字段，预留防作弊配置）。
type StartExamRequest struct {
	ExamID string `json:"exam_id" binding:"required"`
}

// AnswerInput 单题作答输入。
type AnswerInput struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"max=2000"`
}

// SubmitRecordRequest 提交答卷请求。
type SubmitRecordRequest struct {
	Answers     []AnswerInput     `json:"answers" binding:"required,min=1"`
	CheatCount  int               `json:"cheat_count" binding:"omitempty,min=0,max=1000"`
	CheatEvents []CheatEventInput `json:"cheat_events"`
}

// AutoSaveRecordRequest 自动保存草稿请求（断点续答）。
// SaveVersion 为前端本次保存所依据的版本号；服务端只接受比已存版本更新的请求，忽略乱序旧请求。
type AutoSaveRecordRequest struct {
	Answers      []AnswerInput `json:"answers"`
	CurrentIndex int           `json:"current_index" binding:"omitempty,min=0"`
	SaveVersion  int64         `json:"save_version" binding:"required,min=1"`
}

// CheatEventInput 防作弊事件输入。
type CheatEventInput struct {
	Type   string `json:"type" binding:"omitempty,oneof=switch_tab copy_paste blur"`
	Detail string `json:"detail" binding:"omitempty,max=500"`
}

// GradeRequest 主观题批改请求。
type GradeRequest struct {
	Grades []GradeItem `json:"grades" binding:"required,min=1"`
}

// GradeItem 单题批改。
type GradeItem struct {
	QuestionID string  `json:"question_id" binding:"required"`
	Score      float64 `json:"score" binding:"min=0,max=100"`
	Comment    string  `json:"comment" binding:"omitempty,max=500"`
}

// RecordQuery 考试记录查询参数。
type RecordQuery struct {
	ExamID    string `form:"exam_id"`
	Status    string `form:"status"`
	StudentID string `form:"student_id"`
	Page      int64  `form:"page"`
	PageSize  int64  `form:"page_size"`
}

// ExamReportItem 成绩分析单题正确率。
type ExamReportItem struct {
	QuestionID   string  `json:"question_id"`
	Content      string  `json:"content"`
	Type         string  `json:"type"`
	AnswerCount  int     `json:"answer_count"`
	CorrectCount int     `json:"correct_count"`
	Accuracy     float64 `json:"accuracy"`
}

// ExamReport 成绩分析报告。
type ExamReport struct {
	ExamID          string           `json:"exam_id"`
	ExamTitle       string           `json:"exam_title"`
	TotalStudents   int              `json:"total_students"`
	AverageScore    float64          `json:"average_score"`
	MaxScore        float64          `json:"max_score"`
	MinScore        float64          `json:"min_score"`
	PassRate        float64          `json:"pass_rate"`
	ScoreBands      map[string]int   `json:"score_bands"` // 分数段直方图
	QuestionReports []ExamReportItem `json:"question_reports"`
}

// RecordResponse 考试记录响应。
type RecordResponse struct {
	ID              string                  `json:"id"`
	ExamID          string                  `json:"exam_id"`
	ExamTitle       string                  `json:"exam_title"`
	StudentID       string                  `json:"student_id"`
	StudentName     string                  `json:"student_name"`
	Status          string                  `json:"status"`
	StartedAt       time.Time               `json:"started_at"`
	SubmittedAt     *time.Time              `json:"submitted_at"`
	ObjectiveScore  float64                 `json:"objective_score"`
	SubjectiveScore float64                 `json:"subjective_score"`
	FinalScore      float64                 `json:"final_score"`
	PassScore       float64                 `json:"pass_score"`
	CheatCount      int                     `json:"cheat_count"`
	AutoSubmitted   bool                    `json:"auto_submitted"`
	SaveVersion     int64                   `json:"save_version"`
	CurrentIndex    int                     `json:"current_index"`
	LastSavedAt     *time.Time              `json:"last_saved_at"`
	Questions       []model.AttemptQuestion `json:"questions"`
	CreatedAt       time.Time               `json:"created_at"`
}

// ToRecordResponse 模型转响应。
func ToRecordResponse(r *model.ExamRecord) RecordResponse {
	return RecordResponse{
		ID:              r.ID.Hex(),
		ExamID:          r.ExamID.Hex(),
		ExamTitle:       r.ExamTitle,
		StudentID:       r.StudentID.Hex(),
		StudentName:     r.StudentName,
		Status:          r.Status,
		StartedAt:       r.StartedAt,
		SubmittedAt:     r.SubmittedAt,
		ObjectiveScore:  r.ObjectiveScore,
		SubjectiveScore: r.SubjectiveScore,
		FinalScore:      r.FinalScore,
		CheatCount:      r.CheatCount,
		AutoSubmitted:   r.AutoSubmitted,
		SaveVersion:     r.SaveVersion,
		CurrentIndex:    r.CurrentIndex,
		LastSavedAt:     r.LastSavedAt,
		Questions:       r.Questions,
		CreatedAt:       r.CreatedAt,
	}
}
