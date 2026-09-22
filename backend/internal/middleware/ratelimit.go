package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/util"
)

// RateLimit 基于 Redis 的固定窗口限流中间件（每 IP 每分钟限制）。
// Redis 不可用时放行（fail-open），避免限流组件阻断业务。
func RateLimit(rdb *redis.Client, limitPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), time.Now().Format("2006-01-02-15:04"))
		n, err := rdb.Incr(context.Background(), key).Result()
		if err != nil {
			util.Logger.Warn("限流 Redis 不可用，放行", "error", err.Error())
			c.Next()
			return
		}
		if n == 1 {
			_ = rdb.Expire(context.Background(), key, 61*time.Second).Err()
		}
		if n > int64(limitPerMinute) {
			abortWithCode(c, http.StatusTooManyRequests, constants.CodeRateLimited, constants.ErrorCodeText(constants.CodeRateLimited))
			return
		}
		c.Next()
	}
}
