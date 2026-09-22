package router

import (
	"github.com/gin-gonic/gin"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/handler"
	"github.com/onlineexam/onlineexam/internal/middleware"
)

// RegisterWrongBookRoutes 错题本模块路由（学生）。
func RegisterWrongBookRoutes(g *gin.RouterGroup, h *handler.WrongBookHandler) {
	wb := g.Group("/wrong-books", middleware.RequireRoles(constants.RoleStudent))
	{
		wb.GET("", h.List)
		wb.POST("", h.Add)
		wb.GET("/:id", h.Get)
		wb.PUT("/:id", h.Update)
		wb.DELETE("/:id", h.Delete)
	}
}
