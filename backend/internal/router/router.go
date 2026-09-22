// Package router 注册所有路由，统一 /api/v1 前缀与 /healthz。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/onlineexam/onlineexam/internal/config"
	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
	"github.com/onlineexam/onlineexam/internal/service"
)

// Handlers 聚合所有 HTTP 处理器。
type Handlers struct {
	User       *handler.UserHandler
	Question   *handler.QuestionHandler
	Exam       *handler.ExamHandler
	ExamRecord *handler.ExamRecordHandler
	WrongBook  *handler.WrongBookHandler
	Audit      *handler.AuditHandler
}

// Setup 注册全局中间件与全部业务路由。
func Setup(engine *gin.Engine, cfg *config.Config, rdb *redis.Client, hs *Handlers, auditSvc *service.AuditService) {
	engine.Use(
		middleware.RequestID(),
		middleware.ErrorHandler(),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.RateLimit(rdb, cfg.RateLimitPerMinute),
	)

	// 健康检查
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": constants.MsgOK, "data": gin.H{"status": "ok"}})
	})

	api := engine.Group("/api/v1")
	// 审计中间件只挂到写操作组（与业务路由共用）
	api.Use(middleware.Audit(auditSvc))

	// 认证相关
	api.POST("/auth/register", hs.User.Register)
	api.POST("/auth/login", hs.User.Login)

	authGroup := api.Group("")
	authGroup.Use(middleware.Auth(cfg))
	{
		authGroup.GET("/auth/me", hs.User.Me)
		authGroup.PUT("/auth/password", hs.User.ChangePassword)
	}

	RegisterUserRoutes(authGroup, hs.User)
	RegisterQuestionRoutes(authGroup, hs.Question)
	RegisterExamRoutes(authGroup, hs.Exam)
	RegisterExamRecordRoutes(authGroup, hs.ExamRecord)
	RegisterWrongBookRoutes(authGroup, hs.WrongBook)
	RegisterAuditRoutes(authGroup, hs.Audit)
}
