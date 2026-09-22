package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/service"
	"github.com/onlineexam/onlineexam/internal/util"
)

// AuditHandler 审计日志 HTTP 处理器。
type AuditHandler struct {
	svc    *service.AuditService
	logger *slog.Logger
}

// NewAuditHandler 构造审计日志处理器。
func NewAuditHandler(svc *service.AuditService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{svc: svc, logger: logger}
}

// List 分页查询审计日志（管理员）。
func (h *AuditHandler) List(c *gin.Context) {
	var query dto.AuditQuery
	_ = c.ShouldBindQuery(&query)
	filter := bson.M{}
	if query.Module != "" {
		filter["module"] = query.Module
	}
	if query.Action != "" {
		filter["action"] = query.Action
	}
	if query.UserID != "" {
		filter["user_id"] = query.UserID
	}
	if query.Keyword != "" {
		filter["$or"] = bson.A{
			bson.M{"username": bson.M{"$regex": query.Keyword, "$options": "i"}},
			bson.M{"path": bson.M{"$regex": query.Keyword, "$options": "i"}},
		}
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.List(c.Request.Context(), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.AuditResponse, 0, len(list))
	for _, a := range list {
		items = append(items, dto.ToAuditResponse(a))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}
