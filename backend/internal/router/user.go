package router

import (
	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
)

// RegisterUserRoutes 用户模块路由（管理员管理、个人中心）。
func RegisterUserRoutes(g *gin.RouterGroup, h *handler.UserHandler) {
	users := g.Group("/users", middleware.RequireRoles(constants.RoleAdmin))
	{
		users.GET("", h.List)
		users.POST("", h.Create)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.Update)
		users.DELETE("/:id", h.Delete)
	}
}
