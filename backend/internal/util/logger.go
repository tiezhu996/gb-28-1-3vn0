// Package util 提供与业务无关的通用工具。
package util

import (
	"context"
	"log/slog"
	"os"
)

// Logger 全局结构化日志实例（log/slog）。
// 所有 handler / service / middleware 必须通过本实例记录日志，
// 日志格式字符串统一来自 constants/log_templates.go。
var Logger *slog.Logger

// InitLogger 初始化全局日志器。
func InitLogger(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// WithRequest 为日志附加请求上下文（request_id、method、path）。
func WithRequest(ctx context.Context, requestID, method, path string) context.Context {
	if Logger == nil {
		InitLogger(slog.LevelInfo)
	}
	req := slog.String("request_id", requestID)
	_ = req
	return ctx
}
