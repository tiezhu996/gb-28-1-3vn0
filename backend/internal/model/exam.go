package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExamQuestion 试卷中的题目（含该题分值）。
type ExamQuestion struct {
	QuestionID primitive.ObjectID `bson:"question_id" json:"question_id"`
	Score      float64            `bson:"score" json:"score"`
	Order      int                `bson:"order" json:"order"`
}

// Exam 试卷/考试实体，集合 exams。
// 状态枚举：draft / published / ongoing / finished / closed（状态机见 constants/enums.go）。
type Exam struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title          string             `bson:"title" json:"title"`
	Subject        string             `bson:"subject" json:"subject"`
	Description    string             `bson:"description" json:"description"`
	TotalScore     float64            `bson:"total_score" json:"total_score"`
	PassScore      float64            `bson:"pass_score" json:"pass_score"`
	DurationMin    int                `bson:"duration_min" json:"duration_min"`
	StartAt        time.Time          `bson:"start_at" json:"start_at"`
	EndAt          time.Time          `bson:"end_at" json:"end_at"`
	Status         string             `bson:"status" json:"status"`
	ShuffleQuestion bool              `bson:"shuffle_question" json:"shuffle_question"`
	ShuffleOption  bool               `bson:"shuffle_option" json:"shuffle_option"`
	Questions      []ExamQuestion     `bson:"questions" json:"questions"`
	CreatedBy      primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
