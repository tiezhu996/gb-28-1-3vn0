// Package model 定义 MongoDB 集合文档结构。
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户实体，集合 users。
// 角色枚举：admin / teacher / student（见 constants/enums.go）。
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Role         string             `bson:"role" json:"role"`
	Status       string             `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
