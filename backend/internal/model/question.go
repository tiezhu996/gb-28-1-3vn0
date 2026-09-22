package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuestionOption 题目选项。
type QuestionOption struct {
	Key   string `bson:"key" json:"key"`     // A/B/C/D
	Text  string `bson:"text" json:"text"`   // 选项文本
	Score float64 `bson:"score" json:"score"` // 该选项分值（预留）
}

// Question 题目实体，集合 questions。
// 题型枚举：single / multiple / judge / fill / short；难度枚举：easy / medium / hard。
type Question struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type            string             `bson:"type" json:"type"`
	Subject         string             `bson:"subject" json:"subject"`
	KnowledgePoints []string           `bson:"knowledge_points" json:"knowledge_points"`
	Difficulty      string             `bson:"difficulty" json:"difficulty"`
	Content         string             `bson:"content" json:"content"`
	Options         []QuestionOption   `bson:"options,omitempty" json:"options"`
	Answer          string             `bson:"answer" json:"answer"`     // single:A / multiple:A,B / judge:true,false / fill,short:文本
	Analysis        string             `bson:"analysis" json:"analysis"` // 解析
	Score           float64            `bson:"score" json:"score"`       // 默认分值
	Status          string             `bson:"status" json:"status"`     // draft/published
	CreatorID       primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
