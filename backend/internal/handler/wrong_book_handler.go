package handler

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/middleware"
	"github.com/onlineexam/onlineexam/internal/service"
	"github.com/onlineexam/onlineexam/internal/util"
)

// WrongBookHandler 错题本 HTTP 处理器。
type WrongBookHandler struct {
	svc    *service.WrongBookService
	logger *slog.Logger
}

// NewWrongBookHandler 构造错题本处理器。
func NewWrongBookHandler(svc *service.WrongBookService, logger *slog.Logger) *WrongBookHandler {
	return &WrongBookHandler{svc: svc, logger: logger}
}

// Add 加入错题本。
func (h *WrongBookHandler) Add(c *gin.Context) {
	var req dto.AddWrongBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("错题本模块：添加错题参数校验失败（字段 question_id）（%s）", err.Error()), err))
		return
	}
	questionID, err := primitive.ObjectIDFromHex(req.QuestionID)
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "错题本模块：question_id 参数非法"))
		return
	}
	var examID, recordID primitive.ObjectID
	if req.ExamID != "" {
		examID, _ = primitive.ObjectIDFromHex(req.ExamID)
	}
	if req.ExamRecordID != "" {
		recordID, _ = primitive.ObjectIDFromHex(req.ExamRecordID)
	}
	entry, err := h.svc.Add(c.Request.Context(), middleware.GetUserID(c), questionID, examID, recordID, req.Note)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgWrongBookAdded, dto.ToWrongBookResponse(entry))
}

// Update 更新错题本。
func (h *WrongBookHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "错题本模块：id 参数非法"))
		return
	}
	var req dto.UpdateWrongBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("错题本模块：更新参数校验失败（字段 status/note）（%s）", err.Error()), err))
		return
	}
	entry, err := h.svc.Update(c.Request.Context(), id, middleware.GetUserID(c), req.Status, req.Note)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToWrongBookResponse(entry))
}

// Delete 移除错题。
func (h *WrongBookHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "错题本模块：id 参数非法"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, middleware.GetUserID(c)); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Get 查询单个错题。
func (h *WrongBookHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "错题本模块：id 参数非法"))
		return
	}
	entry, err := h.svc.GetByID(c.Request.Context(), id, middleware.GetUserID(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToWrongBookResponse(entry))
}

// List 分页查询错题本。
func (h *WrongBookHandler) List(c *gin.Context) {
	var query dto.WrongBookQuery
	_ = c.ShouldBindQuery(&query)
	filter := bson.M{}
	if query.Subject != "" {
		filter["subject"] = query.Subject
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.KnowledgePoint != "" {
		filter["knowledge_points"] = query.KnowledgePoint
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.List(c.Request.Context(), middleware.GetUserID(c), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.WrongBookResponse, 0, len(list))
	for _, w := range list {
		items = append(items, dto.ToWrongBookResponse(w))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}
