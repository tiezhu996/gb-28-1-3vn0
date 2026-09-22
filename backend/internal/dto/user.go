package dto

import (
	"time"

	"github.com/onlineexam/onlineexam/internal/model"
)

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Role     string `json:"role" binding:"omitempty,oneof=student teacher admin"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// CreateUserRequest 管理员创建用户请求。
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Role     string `json:"role" binding:"required,oneof=student teacher admin"`
}

// UpdateUserRequest 更新用户请求（角色字段校验：admin/teacher/student）。
type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2,max=50"`
	Role   string `json:"role" binding:"omitempty,oneof=student teacher admin"`
	Status string `json:"status" binding:"omitempty,oneof=active disabled"`
}

// ChangePasswordRequest 修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=64"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// UserResponse 用户响应。
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToUserResponse 模型转响应。
func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID.Hex(),
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}
