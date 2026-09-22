package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AttemptOption 考试快照中的选项（已随机打乱）。
type AttemptOption struct {
	Key  string `bson:"key" json:"key"`
	Text string `bson:"text" json:"text"`
}

// AttemptQuestion 考试快照中的题目。
type AttemptQuestion struct {
	QuestionID      primitive.ObjectID `bson:"question_id" json:"question_id"`
	Type            string             `bson:"type" json:"type"`
	Subject         string             `bson:"subject" json:"subject"`
	KnowledgePoints []string           `bson:"knowledge_points" json:"knowledge_points"`
	Content         string             `bson:"content" json:"content"`
	Options         []AttemptOption    `bson:"options,omitempty" json:"options"`
	Score           float64            `bson:"score" json:"score"`
	CorrectAnswer   string             `bson:"correct_answer" json:"correct_answer"` // 客观题标准答案快照
	UserAnswer      string             `bson:"user_answer" json:"user_answer"`       // 学生作答
	Result          string             `bson:"result" json:"result"`                 // correct/wrong/partial/unmarked
	GotScore        float64            `bson:"got_score" json:"got_score"`           // 客观题自动得分
	SubjectiveScore float64            `bson:"subjective_score,omitempty" json:"subjective_score"`
	Comment         string             `bson:"comment,omitempty" json:"comment"`
	Marked          bool               `bson:"marked" json:"marked"` // 教师是否已批改
}

// CheatEvent 切屏/防作弊事件。
type CheatEvent struct {
	Type       string    `bson:"type" json:"type"` // switch_tab / copy_paste / blur
	Detail     string    `bson:"detail" json:"detail"`
	OccurredAt time.Time `bson:"occurred_at" json:"occurred_at"`
}

// ExamRecord 考试记录/答卷实体，集合 exam_records。
// 状态枚举：in_progress / submitted / graded；结果枚举：correct / wrong / partial / unmarked。
type ExamRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExamID          primitive.ObjectID `bson:"exam_id" json:"exam_id"`
	ExamTitle       string             `bson:"exam_title" json:"exam_title"`
	StudentID       primitive.ObjectID `bson:"student_id" json:"student_id"`
	StudentName     string             `bson:"student_name" json:"student_name"`
	Questions       []AttemptQuestion  `bson:"questions" json:"questions"`
	Status          string             `bson:"status" json:"status"`
	StartedAt       time.Time          `bson:"started_at" json:"started_at"`
	SubmittedAt     *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at"`
	ObjectiveScore  float64            `bson:"objective_score" json:"objective_score"` // 客观题自动得分
	SubjectiveScore float64            `bson:"subjective_score" json:"subjective_score"`
	FinalScore      float64            `bson:"final_score" json:"final_score"` // 最终总分
	CheatCount      int                `bson:"cheat_count" json:"cheat_count"`
	CheatEvents     []CheatEvent       `bson:"cheat_events,omitempty" json:"cheat_events"`
	AutoSubmitted   bool               `bson:"auto_submitted" json:"auto_submitted"`
	SaveVersion     int64              `bson:"save_version" json:"save_version"`             // 自动保存版本号（单调递增，乱序旧请求保护）
	CurrentIndex    int                `bson:"current_index" json:"current_index"`           // 断点续答：当前题号
	LastSavedAt     *time.Time         `bson:"last_saved_at,omitempty" json:"last_saved_at"` // 最后一次自动保存时间
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
