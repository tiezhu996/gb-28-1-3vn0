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

// ExamRecordHandler 考试记录 HTTP 处理器。
type ExamRecordHandler struct {
	svc    *service.ExamRecordService
	logger *slog.Logger
}

// NewExamRecordHandler 构造考试记录处理器。
func NewExamRecordHandler(svc *service.ExamRecordService, logger *slog.Logger) *ExamRecordHandler {
	return &ExamRecordHandler{svc: svc, logger: logger}
}

// Start 学生开始考试。
func (h *ExamRecordHandler) Start(c *gin.Context) {
	examID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：examId 参数非法"))
		return
	}
	studentID := middleware.GetUserID(c)
	claims := middleware.GetClaims(c)
	name := ""
	if claims != nil {
		name = claims.Name
	}
	rec, err := h.svc.StartExam(c.Request.Context(), examID, studentID, name)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToRecordResponse(rec))
}

// Submit 学生提交答卷。
func (h *ExamRecordHandler) Submit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：id 参数非法"))
		return
	}
	var req dto.SubmitRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("考试记录模块：提交答卷参数校验失败（字段 answers）（%s）", err.Error()), err))
		return
	}
	rec, err := h.svc.Submit(c.Request.Context(), id, req.Answers, req.CheatCount, req.CheatEvents, false)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgRecordSubmitSuccess, dto.ToRecordResponse(rec))
}

// SaveDraft 学生自动保存作答草稿（断点续答）。
func (h *ExamRecordHandler) SaveDraft(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：id 参数非法"))
		return
	}
	var req dto.SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("考试记录模块：草稿保存参数校验失败（%s）", err.Error()), err))
		return
	}
	res, err := h.svc.SaveDraft(c.Request.Context(), id, middleware.GetUserID(c), req.Answers, req.CurrentIndex, req.Version)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// AutoSubmit 超时自动提交（教师/管理员触发或定时任务）。
func (h *ExamRecordHandler) AutoSubmit(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：id 参数非法"))
		return
	}
	rec, err := h.svc.Submit(c.Request.Context(), id, nil, 0, nil, true)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToRecordResponse(rec))
}

// Grade 教师批改主观题。
func (h *ExamRecordHandler) Grade(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：id 参数非法"))
		return
	}
	var req dto.GradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("考试记录模块：批改参数校验失败（字段 grades）（%s）", err.Error()), err))
		return
	}
	rec, err := h.svc.Grade(c.Request.Context(), id, req.Grades, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgRecordGradedSuccess, dto.ToRecordResponse(rec))
}

// Get 查询单个考试记录。
func (h *ExamRecordHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：id 参数非法"))
		return
	}
	rec, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToRecordResponse(rec))
}

// ListMine 学生查询自己的考试记录。
func (h *ExamRecordHandler) ListMine(c *gin.Context) {
	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.ListByStudent(c.Request.Context(), middleware.GetUserID(c), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.RecordResponse, 0, len(list))
	for _, r := range list {
		items = append(items, dto.ToRecordResponse(r))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}

// ListByExam 教师按试卷查询考试记录。
func (h *ExamRecordHandler) ListByExam(c *gin.Context) {
	examID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：examId 参数非法"))
		return
	}
	filter := bson.M{}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.ListByExam(c.Request.Context(), examID, filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.RecordResponse, 0, len(list))
	for _, r := range list {
		items = append(items, dto.ToRecordResponse(r))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}

// Report 成绩分析报告。
func (h *ExamRecordHandler) Report(c *gin.Context) {
	examID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "考试记录模块：examId 参数非法"))
		return
	}
	report, err := h.svc.Report(c.Request.Context(), examID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, report)
}
