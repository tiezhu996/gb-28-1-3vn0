package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/util"
)

// RequireRoles RBAC 权限中间件：仅允许指定角色访问（admin/teacher/student）。
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "" {
			abortWithCode(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgUnauthorized)
			return
		}
		if _, ok := allowed[role]; !ok {
			util.Logger.Warn("RBAC 拦截", "role", role, "path", c.FullPath(), "roles", roles)
			abortWithCode(c, http.StatusForbidden, constants.CodeForbidden, constants.MsgForbidden)
			return
		}
		c.Next()
	}
}
