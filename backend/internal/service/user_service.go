package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/config"
	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
	"github.com/onlineexam/onlineexam/internal/util"
)

// UserService 用户服务：注册、登录、用户管理、种子数据。
type UserService struct {
	repo   repository.UserRepository
	logger *slog.Logger
	cfg    *config.Config
}

// NewUserService 构造用户服务。
func NewUserService(repo repository.UserRepository, logger *slog.Logger, cfg *config.Config) *UserService {
	return &UserService{repo: repo, logger: logger, cfg: cfg}
}

// Register 学生/教师注册。
func (s *UserService) Register(ctx context.Context, name, email, password, role string) (*model.User, error) {
	if !constants.IsValidUserRole(role) {
		return nil, util.NewAppError(constants.CodeUserInvalidRole, fmt.Sprintf(constants.MsgValidationFailed, "role")+fmt.Sprintf("（角色 %s 非法）", role))
	}
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user service register: %w", err)
	}
	if exists {
		return nil, util.NewAppError(constants.CodeUserEmailExists, fmt.Sprintf(constants.MsgUserEmailExists, email))
	}
	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("user service register hash: %w", err)
	}
	now := time.Now()
	u := &model.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Status:       constants.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("user service register create: %w", err)
	}
	s.logger.Info(constants.LogUserRegistered, "role", u.Role, "email", u.Email)
	return u, nil
}

// Login 登录，返回用户与 JWT。
func (s *UserService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", util.NewAppError(constants.CodeUserNotFound, fmt.Sprintf(constants.MsgUserNotFound, email))
		}
		return nil, "", fmt.Errorf("user service login find: %w", err)
	}
	if u.Status == constants.UserStatusDisabled {
		s.logger.Warn(constants.LogUserLoginFailed, "role", u.Role, "email", email, "error", "disabled")
		return nil, "", util.NewAppError(constants.CodeUserDisabled, fmt.Sprintf(constants.MsgUserDisabled, u.Role))
	}
	if !util.CheckPassword(u.PasswordHash, password) {
		s.logger.Warn(constants.LogUserLoginFailed, "role", u.Role, "email", email, "error", "bad password")
		return nil, "", util.NewAppError(constants.CodeUserBadPassword, constants.MsgUserBadPassword)
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, u.ID.Hex(), u.Email, u.Role, u.Name, s.cfg.JWTExpiresMinutes)
	if err != nil {
		return nil, "", fmt.Errorf("user service login token: %w", err)
	}
	s.logger.Info(constants.LogUserLogin, "role", u.Role, "email", u.Email)
	return u, token, nil
}

// Create 管理员创建用户。
func (s *UserService) Create(ctx context.Context, name, email, password, role, operator string) (*model.User, error) {
	if !constants.IsValidUserRole(role) {
		return nil, util.NewAppError(constants.CodeUserInvalidRole, fmt.Sprintf(constants.MsgValidationFailed, "role"))
	}
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user service create: %w", err)
	}
	if exists {
		return nil, util.NewAppError(constants.CodeUserEmailExists, fmt.Sprintf(constants.MsgUserEmailExists, email))
	}
	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("user service create hash: %w", err)
	}
	now := time.Now()
	u := &model.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		Status:       constants.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("user service create: %w", err)
	}
	s.logger.Info(constants.LogUserCreated, "role", u.Role, "email", u.Email, "operator", operator)
	return u, nil
}

// Update 更新用户（角色/状态）。
func (s *UserService) Update(ctx context.Context, id primitive.ObjectID, name, role, status, operator string) (*model.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeUserNotFound, fmt.Sprintf(constants.MsgUserNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("user service update find: %w", err)
	}
	if name != "" {
		u.Name = name
	}
	if role != "" {
		if !constants.IsValidUserRole(role) {
			return nil, util.NewAppError(constants.CodeUserInvalidRole, fmt.Sprintf(constants.MsgValidationFailed, "role"))
		}
		u.Role = role
	}
	if status != "" {
		if status != constants.UserStatusActive && status != constants.UserStatusDisabled {
			return nil, util.NewAppError(constants.CodeBadRequest, fmt.Sprintf(constants.MsgValidationFailed, "status"))
		}
		u.Status = status
	}
	u.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("user service update: %w", err)
	}
	s.logger.Info(constants.LogUserUpdated, "role", u.Role, "email", u.Email, "operator", operator)
	return u, nil
}

// Delete 删除用户。
func (s *UserService) Delete(ctx context.Context, id primitive.ObjectID, operator string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeUserNotFound, fmt.Sprintf(constants.MsgUserNotFound, id.Hex()))
		}
		return fmt.Errorf("user service delete: %w", err)
	}
	s.logger.Info(constants.LogUserDeleted, "email", id.Hex(), "operator", operator)
	return nil
}

// ChangePassword 修改密码。
func (s *UserService) ChangePassword(ctx context.Context, id primitive.ObjectID, oldPwd, newPwd string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeUserNotFound, fmt.Sprintf(constants.MsgUserNotFound, id.Hex()))
		}
		return fmt.Errorf("user service change password find: %w", err)
	}
	if !util.CheckPassword(u.PasswordHash, oldPwd) {
		return util.NewAppError(constants.CodeUserBadPassword, constants.MsgUserBadPassword)
	}
	hash, err := util.HashPassword(newPwd)
	if err != nil {
		return fmt.Errorf("user service change password hash: %w", err)
	}
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, u); err != nil {
		return fmt.Errorf("user service change password: %w", err)
	}
	s.logger.Info(constants.LogUserPasswordReset, "email", u.Email, "operator", id.Hex())
	return nil
}

// List 分页查询用户。
func (s *UserService) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.User, int64, error) {
	list, total, err := s.repo.List(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("user service list: %w", err)
	}
	return list, total, nil
}

// GetByID 查询单个用户。
func (s *UserService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeUserNotFound, fmt.Sprintf(constants.MsgUserNotFound, id.Hex()))
		}
		return nil, fmt.Errorf("user service get: %w", err)
	}
	return u, nil
}

// Seed 初始化种子用户（admin/teacher/student）。
func (s *UserService) Seed(ctx context.Context) error {
	seeds := []struct {
		name, email, password, role string
	}{
		{"系统管理员", "admin@onlineexam.com", "admin123456", constants.RoleAdmin},
		{"张老师", "teacher@onlineexam.com", "teacher123456", constants.RoleTeacher},
		{"李同学", "student@onlineexam.com", "student123456", constants.RoleStudent},
	}
	for _, sd := range seeds {
		exists, err := s.repo.ExistsByEmail(ctx, sd.email)
		if err != nil {
			return fmt.Errorf("user service seed: %w", err)
		}
		if exists {
			continue
		}
		if _, err := s.Register(ctx, sd.name, sd.email, sd.password, sd.role); err != nil {
			return fmt.Errorf("user service seed register: %w", err)
		}
	}
	return nil
}
