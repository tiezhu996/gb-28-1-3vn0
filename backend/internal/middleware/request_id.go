// Package middleware 提供认证、RBAC、请求追踪、审计、限流等中间件。
package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const (
	// RequestIDKey gin context 中请求 ID 的键。
	RequestIDKey = "request_id"
)

// RequestID 为每个请求注入唯一 request_id，并回写响应头 X-Request-Id。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = newRequestID()
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

// GetRequestID 从上下文读取 request_id。
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newRequestID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
