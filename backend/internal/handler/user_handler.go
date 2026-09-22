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

// UserHandler 用户 HTTP 处理器。
type UserHandler struct {
	svc    *service.UserService
	logger *slog.Logger
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// Register 注册（默认学生/教师）。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("用户模块：注册参数校验失败（字段 name/email/password/role）（%s）", err.Error()), err))
		return
	}
	role := req.Role
	if role == "" {
		role = constants.RoleStudent
	}
	u, err := h.svc.Register(c.Request.Context(), req.Name, req.Email, req.Password, role)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgRegisterSuccess, dto.ToUserResponse(u))
}

// Login 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("用户模块：登录参数校验失败（字段 email/password）（%s）", err.Error()), err))
		return
	}
	u, token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessMessage(c, constants.MsgLoginSuccess, dto.LoginResponse{Token: token, User: dto.ToUserResponse(u)})
}

// Me 当前登录用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	id := middleware.GetUserID(c)
	if id.IsZero() {
		Error(c, util.NewAppError(constants.CodeUnauthorized, constants.MsgUnauthorized))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToUserResponse(u))
}

// List 分页查询用户（管理员）。
func (h *UserHandler) List(c *gin.Context) {
	filter := bson.M{}
	if role := c.Query("role"); role != "" {
		filter["role"] = role
	}
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if kw := c.Query("keyword"); kw != "" {
		filter["$or"] = bson.A{
			bson.M{"name": bson.M{"$regex": kw, "$options": "i"}},
			bson.M{"email": bson.M{"$regex": kw, "$options": "i"}},
		}
	}
	page := util.GetPageParams(c, 20)
	list, total, err := h.svc.List(c.Request.Context(), filter, page.Page, page.PageSize)
	if err != nil {
		Error(c, err)
		return
	}
	items := make([]dto.UserResponse, 0, len(list))
	for _, u := range list {
		items = append(items, dto.ToUserResponse(u))
	}
	PageResult(c, items, total, page.Page, page.PageSize)
}

// Create 管理员创建用户。
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("用户模块：创建用户参数校验失败（字段 name/email/password/role）（%s）", err.Error()), err))
		return
	}
	u, err := h.svc.Create(c.Request.Context(), req.Name, req.Email, req.Password, req.Role, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToUserResponse(u))
}

// Update 更新用户。
func (h *UserHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "用户模块：id 参数非法"))
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("用户模块：更新用户参数校验失败（字段 name/role/status）（%s）", err.Error()), err))
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, req.Name, req.Role, req.Status, middleware.GetEmail(c))
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToUserResponse(u))
}

// Delete 删除用户。
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "用户模块：id 参数非法"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id, middleware.GetEmail(c)); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}

// GetUser 查询单个用户（管理员）。
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		Error(c, util.NewAppError(constants.CodeBadRequest, "用户模块：id 参数非法"))
		return
	}
	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto.ToUserResponse(u))
}

// ChangePassword 修改密码。
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, util.WrapAppError(constants.CodeValidationFailed, fmt.Sprintf("用户模块：修改密码参数校验失败（字段 old_password/new_password）（%s）", err.Error()), err))
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), middleware.GetUserID(c), req.OldPassword, req.NewPassword); err != nil {
		Error(c, err)
		return
	}
	Success(c, nil)
}
