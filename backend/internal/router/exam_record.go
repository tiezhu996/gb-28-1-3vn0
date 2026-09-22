package router

import (
	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
)

// RegisterExamRecordRoutes 考试记录模块路由。
func RegisterExamRecordRoutes(g *gin.RouterGroup, h *handler.ExamRecordHandler) {
	// 学生：开始考试 / 我的记录 / 提交
	student := g.Group("/exam-records", middleware.RequireRoles(constants.RoleStudent))
	{
		student.POST("/:id/start", h.Start)
		student.GET("/mine", h.ListMine)
		student.POST("/:id/submit", h.Submit)
		student.POST("/:id/save", h.SaveDraft)
	}

	// 教师/管理员：按试卷查记录、批改、报告
	teacher := g.Group("/exam-records", middleware.RequireRoles(constants.RoleTeacher, constants.RoleAdmin))
	{
		teacher.GET("/exam/:id", h.ListByExam)
		teacher.POST("/:id/grade", h.Grade)
		teacher.POST("/:id/auto-submit", h.AutoSubmit)
	}
	g.GET("/exams/:id/records", middleware.RequireRoles(constants.RoleTeacher, constants.RoleAdmin), h.ListByExam)
	g.GET("/exams/:id/report", middleware.RequireRoles(constants.RoleTeacher, constants.RoleAdmin), h.Report)
	g.GET("/exam-records/:id", h.Get)
}
