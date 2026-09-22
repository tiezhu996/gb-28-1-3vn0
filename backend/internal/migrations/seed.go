package migrations

import (
	"context"
	"fmt"

	"github.com/onlineexam/onlineexam/internal/service"
)

// Seed 初始化种子数据（admin/teacher/student 账号）。
func Seed(ctx context.Context, userSvc *service.UserService) error {
	if err := userSvc.Seed(ctx); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	return nil
}
