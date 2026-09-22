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

// ExamHandler 试卷 HTTP 处理器。
type ExamHandler struct {
	svc    *service.ExamService
	logger *slog.Logger
}

// NewExamHandler 构造试卷处理器。
func NewExamHandler(svc *service.ExamService, logger *slog.Logger) *ExamHandler {
	return &ExamHandler{svc: svc, logger: logger}
}

// Create 手动创建试卷。
func (h *ExamHandler) Create(c *gin.Context) {
	var req dto.CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("试卷模块：创建试卷参数校验失败（字段 title/subject/total_score/duration_min/start_at/end_at）（%s）", err.Error()), err))
		return
	}
	exam, err := h.svc.Create(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToExamResponse(exam))
}

// AutoGenerate 自动组卷。
func (h *ExamHandler) AutoGenerate(c *gin.Context) {
	var req dto.AutoGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("试卷模块：自动组卷参数校验失败（字段 title/subject/knowledge_points/difficulty_dist/score_per_question）（%s）", err.Error()), err))
		return
	}
	exam, err := h.svc.AutoGenerate(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgExamAutoGenerate, dto.ToExamResponse(exam))
}

// Update 更新试卷。
func (h *ExamHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "试卷模块：id 参数非法"))
		return
	}
	var req dto.UpdateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("试卷模块：更新试卷参数校验失败（%s）", err.Error()), err))
		return
	}
	exam, err := h.svc.Update(c.Request.Context(), id, &req, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToExamResponse(exam))
}

// Publish 发布试卷。
func (h *ExamHandler) Publish(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "试卷模块：id 参数非法"))
		return
	}
	exam, err := h.svc.Publish(c.Request.Context(), id, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgExamPublishSuccess, dto.ToExamResponse(exam))
}

// Close 关闭试卷。
func (h *ExamHandler) Close(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "试卷模块：id 参数非法"))
		return
	}
	exam, err := h.svc.Close(c.Request.Context(), id, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToExamResponse(exam))
}

// Get 查询单个试卷。
func (h *ExamHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "试卷模块：id 参数非法"))
		return
	}
	exam, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToExamResponse(exam))
}

// List 分页查询试卷。
func (h *ExamHandler) List(c *gin.Context) {
	var query dto.ExamQuery
	_ = c.ShouldBindQuery(&query)
	filter := bson.M{}
	if query.Title != "" {
		filter["title"] = bson.M{"$regex": query.Title, "$options": "i"}
	}
	if query.Subject != "" {
		filter["subject"] = query.Subject
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.List(c.Request.Context(), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.ExamResponse, 0, len(list))
	for _, e := range list {
		items = append(items, dto.ToExamResponse(e))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}

// Delete 删除试卷。
func (h *ExamHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "试卷模块：id 参数非法"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, middleware.GetEmail(c)); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
