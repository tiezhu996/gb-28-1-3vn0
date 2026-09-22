package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog 操作审计日志实体，集合 audit_logs。
type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Role      string             `bson:"role" json:"role"`
	Module    string             `bson:"module" json:"module"`
	Action    string             `bson:"action" json:"action"`
	Method    string             `bson:"method" json:"method"`
	Path      string             `bson:"path" json:"path"`
	StatusCode int               `bson:"status_code" json:"status_code"`
	RequestID string             `bson:"request_id" json:"request_id"`
	ClientIP  string             `bson:"client_ip" json:"client_ip"`
	Detail    string             `bson:"detail" json:"detail"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
