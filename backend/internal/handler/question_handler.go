package handler

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/dto"
	"github.com/onlineexam/onlineexam/internal/middleware"
	"github.com/onlineexam/onlineexam/internal/service"
	"github.com/onlineexam/onlineexam/internal/util"
)

// QuestionHandler 题库 HTTP 处理器。
type QuestionHandler struct {
	svc    *service.QuestionService
	logger *slog.Logger
}

// NewQuestionHandler 构造题库处理器。
func NewQuestionHandler(svc *service.QuestionService, logger *slog.Logger) *QuestionHandler {
	return &QuestionHandler{svc: svc, logger: logger}
}

// Create 创建题目。
func (h *QuestionHandler) Create(c *gin.Context) {
	var req dto.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("题库模块：创建题目参数校验失败（字段 type/subject/knowledge_points/difficulty/content/answer/score）（%s）", err.Error()), err))
		return
	}
	q, err := h.svc.Create(c.Request.Context(), &req, middleware.GetUserID(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToQuestionResponse(q))
}

// Update 更新题目。
func (h *QuestionHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "题库模块：id 参数非法"))
		return
	}
	var req dto.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("题库模块：更新题目参数校验失败（%s）", err.Error()), err))
		return
	}
	q, err := h.svc.Update(c.Request.Context(), id, &req, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToQuestionResponse(q))
}

// Delete 删除题目。
func (h *QuestionHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "题库模块：id 参数非法"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, middleware.GetEmail(c)); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// Get 查询单个题目。
func (h *QuestionHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "题库模块：id 参数非法"))
		return
	}
	q, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToQuestionResponse(q))
}

// List 分页查询题目。
func (h *QuestionHandler) List(c *gin.Context) {
	var query dto.QuestionQuery
	_ = c.ShouldBindQuery(&query)
	filter := bson.M{}
	if query.Subject != "" {
		filter["subject"] = query.Subject
	}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.Difficulty != "" {
		filter["difficulty"] = query.Difficulty
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.KnowledgePoint != "" {
		filter["knowledge_points"] = query.KnowledgePoint
	}
	if query.Keyword != "" {
		filter["content"] = bson.M{"$regex": query.Keyword, "$options": "i"}
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.List(c.Request.Context(), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.QuestionResponse, 0, len(list))
	for _, q := range list {
		items = append(items, dto.ToQuestionResponse(q))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}

// Import 批量导入题目（Excel 模板上传）。
func (h *QuestionHandler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "题库模块：请上传 file 字段的 Excel 文件"))
		return
	}
	src, err := file.Open()
	if err != nil {
		Error(c, util.NewAppError(constants.CodeQuestionImportErr, "题库模块：读取上传文件失败"))
		return
	}
	defer func() { _ = src.Close() }()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		Error(c, util.NewAppError(constants.CodeQuestionImportErr, "题库模块：读取上传文件失败"))
		return
	}
	rows, err := util.ParseQuestionExcel(buf.Bytes())
	if err != nil {
		Error(c, err)
		return
	}
	count, err := h.svc.Import(c.Request.Context(), rows, middleware.GetUserID(c))
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgQuestionImportSuccess, gin.H{"count": count})
}

// Template 下载题目导入 Excel 模板。
func (h *QuestionHandler) Template(c *gin.Context) {
	f := excelize.NewFile()
	sheet := "Sheet1"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"题型", "学科", "知识点(逗号分隔)", "难度", "题干", "选项A", "选项B", "选项C", "选项D", "答案", "解析", "分值"}
	for i, hd := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, hd)
	}
	example := []interface{}{"single", "计算机基础", "数据结构", "easy", "以下哪种结构是先进后出？", "队列", "栈", "数组", "链表", "B", "栈是先进后出结构", "5"}
	for i, v := range example {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		_ = f.SetCellValue(sheet, cell, v)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		Error(c, util.NewAppError(constants.CodeQuestionImportErr, "题库模块：生成模板失败"))
		return
	}
	c.Header("Content-Disposition", "attachment; filename=question_template.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
