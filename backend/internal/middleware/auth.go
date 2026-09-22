package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/config"
	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/util"
)

const (
	// ClaimsKey gin context 中 JWT 载荷的键。
	ClaimsKey = "auth_claims"
)

// Auth 认证中间件：校验 Authorization: Bearer <token>，解析 JWT 后写入上下文。
func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := util.ParseToken(cfg.JWTSecret, token)
		if err != nil {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			return
		}
		if _, err := primitive.ObjectIDFromHex(claims.UserID); err != nil {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			return
		}
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// GetClaims 从上下文读取 JWT 载荷。
func GetClaims(c *gin.Context) *util.Claims {
	if v, ok := c.Get(ClaimsKey); ok {
		if claims, ok := v.(*util.Claims); ok {
			return claims
		}
	}
	return nil
}

// GetUserID 从上下文读取当前用户 ID。
func GetUserID(c *gin.Context) primitive.ObjectID {
	claims := GetClaims(c)
	if claims == nil {
		return primitive.NilObjectID
	}
	oid, _ := primitive.ObjectIDFromHex(claims.UserID)
	return oid
}

// GetRole 从上下文读取当前用户角色。
func GetRole(c *gin.Context) string {
	claims := GetClaims(c)
	if claims == nil {
		return ""
	}
	return claims.Role
}

// GetEmail 从上下文读取当前用户邮箱。
func GetEmail(c *gin.Context) string {
	claims := GetClaims(c)
	if claims == nil {
		return ""
	}
	return claims.Email
}

func abortWithCode(c *gin.Context, status, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message, "data": nil})
}
