package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// AuditQuery 审计日志查询参数。
type AuditQuery struct {
	Module  string `form:"module"`
	Action  string `form:"action"`
	UserID  string `form:"user_id"`
	Keyword string `form:"keyword"`
	Page    int64  `form:"page"`
	PageSize int64 `form:"page_size"`
}

// AuditResponse 审计日志响应。
type AuditResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	Module     string    `json:"module"`
	Action     string    `json:"action"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	RequestID  string    `json:"request_id"`
	ClientIP   string    `json:"client_ip"`
	Detail     string    `json:"detail"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToAuditResponse 模型转响应。
func ToAuditResponse(a *model.AuditLog) AuditResponse {
	return AuditResponse{
		ID:         a.ID.Hex(),
		UserID:     a.UserID.Hex(),
		Username:   a.Username,
		Role:       a.Role,
		Module:     a.Module,
		Action:     a.Action,
		Method:     a.Method,
		Path:       a.Path,
		StatusCode: a.StatusCode,
		RequestID:  a.RequestID,
		ClientIP:   a.ClientIP,
		Detail:     a.Detail,
		CreatedAt:  a.CreatedAt,
	}
}
