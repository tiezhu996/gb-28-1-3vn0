package router

import (
	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
)

// RegisterAuditRoutes 审计日志路由（管理员）。
func RegisterAuditRoutes(g *gin.RouterGroup, h *handler.AuditHandler) {
	audit := g.Group("/audit-logs", middleware.RequireRoles(constants.RoleAdmin))
	{
		audit.GET("", h.List)
	}
}
