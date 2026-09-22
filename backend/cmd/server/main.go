// Package main 在线考试系统后端入口。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/config"
	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/database"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/migrations"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/router"
	"github.com/onlineexam/onlineexam/internal/service"
	"github.com/onlineexam/onlineexam/internal/util"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err.Error())
		os.Exit(1)
	}
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	util.InitLogger(level)
	util.Logger.Info(constants.LogServerStarting, "port", cfg.ServerPort)

	gin.SetMode(cfg.RunMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		util.Logger.Error("连接数据库失败", "error", err.Error())
		os.Exit(1)
	}
	defer func() { _ = db.Close(ctx) }()

	if err := migrations.EnsureIndexes(ctx, db.DB); err != nil {
		util.Logger.Error("初始化索引失败", "error", err.Error())
		os.Exit(1)
	}

	// 装配依赖（构造器注入，单向依赖 handler→service→repository→model）
	userRepo := repository.NewMongoUserRepository(db.DB)
	questionRepo := repository.NewMongoQuestionRepository(db.DB)
	examRepo := repository.NewMongoExamRepository(db.DB)
	recordRepo := repository.NewMongoExamRecordRepository(db.DB)
	wrongBookRepo := repository.NewMongoWrongBookRepository(db.DB)
	auditRepo := repository.NewMongoAuditRepository(db.DB)

	userSvc := service.NewUserService(userRepo, util.Logger, cfg)
	questionSvc := service.NewQuestionService(questionRepo, util.Logger)
	examSvc := service.NewExamService(examRepo, questionSvc, util.Logger)
	recordSvc := service.NewExamRecordService(recordRepo, examSvc, util.Logger)
	wrongBookSvc := service.NewWrongBookService(wrongBookRepo, questionSvc, recordSvc, util.Logger)
	auditSvc := service.NewAuditService(auditRepo, util.Logger)

	if cfg.SeedEnabled {
		if err := migrations.Seed(ctx, userSvc); err != nil {
			util.Logger.Warn("种子数据初始化失败", "error", err.Error())
		}
	}

	hs := &router.Handlers{
		User:       handler.NewUserHandler(userSvc, util.Logger),
		Question:   handler.NewQuestionHandler(questionSvc, util.Logger),
		Exam:       handler.NewExamHandler(examSvc, util.Logger),
		ExamRecord: handler.NewExamRecordHandler(recordSvc, util.Logger),
		WrongBook:  handler.NewWrongBookHandler(wrongBookSvc, util.Logger),
		Audit:      handler.NewAuditHandler(auditSvc, util.Logger),
	}

	engine := gin.New()
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/healthz"}}))
	router.Setup(engine, cfg, db.Redis, hs, auditSvc)

	srv := &http.Server{
		Addr:    cfg.ServerHost + ":" + cfg.ServerPort,
		Handler: engine,
	}

	// 后台定时自动提交超时答卷
	go autoSubmitLoop(ctx, recordSvc)

	go func() {
		util.Logger.Info(constants.LogServerStarted, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			util.Logger.Error("服务启动失败", "error", err.Error())
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	util.Logger.Info(constants.LogServerShutdown)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func autoSubmitLoop(ctx context.Context, recordSvc *service.ExamRecordService) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			count, err := recordSvc.AutoSubmitExpired(ctx, now)
			if err != nil {
				util.Logger.Warn("定时自动提交失败", "error", err.Error())
				continue
			}
			if count > 0 {
				util.Logger.Info("定时自动提交完成", "count", count)
			}
		}
	}
}
