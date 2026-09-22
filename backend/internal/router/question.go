package router

import (
	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
)

// RegisterQuestionRoutes 题库模块路由（教师/管理员维护，学生只读）。
func RegisterQuestionRoutes(g *gin.RouterGroup, h *handler.QuestionHandler) {
	// 模板下载对所有登录用户开放
	g.GET("/questions/template", h.Template)
	// 查询对学生、教师、管理员开放
	g.GET("/questions", h.List)
	g.GET("/questions/:id", h.Get)

	write := g.Group("/questions", middleware.RequireRoles(constants.RoleTeacher, constants.RoleAdmin))
	{
		write.POST("", h.Create)
		write.POST("/import", h.Import)
		write.PUT("/:id", h.Update)
		write.DELETE("/:id", h.Delete)
	}
}
