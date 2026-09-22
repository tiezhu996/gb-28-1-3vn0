package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/service"
	"github.com/onlineexam/onlineexam/internal/util"
)

// Audit 操作审计中间件：对写操作（POST/PUT/PATCH/DELETE）记录审计日志。
// 日志落库到 audit_logs 集合，管理员可在审计页面查看。
func Audit(auditSvc *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		method := c.Request.Method
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			return
		}
		// 登录/注册单独记录
		if strings.HasSuffix(c.FullPath(), "/auth/login") || strings.HasSuffix(c.FullPath(), "/auth/register") {
			return
		}
		claims := GetClaims(c)
		userID := primitive.NilObjectID
		username := ""
		role := ""
		if claims != nil {
			userID, _ = primitive.ObjectIDFromHex(claims.UserID)
			username = claims.Name
			role = claims.Role
		}
		entry := &model.AuditLog{
			ID:         primitive.NewObjectID(),
			UserID:     userID,
			Username:   username,
			Role:       role,
			Module:     moduleFromPath(c.FullPath()),
			Action:     actionFromMethod(method),
			Method:     method,
			Path:       c.FullPath(),
			StatusCode: c.Writer.Status(),
			RequestID:  GetRequestID(c),
			ClientIP:   c.ClientIP(),
			Detail:     c.Request.URL.RawQuery,
			CreatedAt:  time.Now(),
		}
		if err := auditSvc.Create(c.Copy().Request.Context(), entry); err != nil {
			util.Logger.Warn("审计日志写入失败", "error", err.Error())
		}
	}
}

func moduleFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

func actionFromMethod(method string) string {
	switch method {
	case "POST":
		return constants.AuditActionCreate
	case "PUT", "PATCH":
		return constants.AuditActionUpdate
	case "DELETE":
		return constants.AuditActionDelete
	default:
		return "unknown"
	}
}
