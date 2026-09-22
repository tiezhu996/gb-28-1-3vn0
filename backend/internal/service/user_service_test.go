package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/onlineexam/onlineexam/internal/config"
	"github.com/onlineexam/onlineexam/internal/constants"
	"github.com/onlineexam/onlineexam/internal/model"
	"github.com/onlineexam/onlineexam/internal/repository"
)

// fakeUserRepo 内存版用户仓储（表驱动测试用）。
type fakeUserRepo struct {
	users map[string]*model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*model.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, u *model.User) error {
	if _, ok := f.users[u.Email]; ok {
		return errors.Join(repository.ErrConflict, errors.New("duplicate email"))
	}
	f.users[u.Email] = u
	return nil
}
func (f *fakeUserRepo) Update(_ context.Context, u *model.User) error {
	if _, ok := f.users[u.Email]; !ok {
		return repository.ErrNotFound
	}
	f.users[u.Email] = u
	return nil
}
func (f *fakeUserRepo) Delete(_ context.Context, id primitive.ObjectID) error {
	for k, u := range f.users {
		if u.ID == id {
			delete(f.users, k)
			return nil
		}
	}
	return repository.ErrNotFound
}
func (f *fakeUserRepo) FindByID(_ context.Context, id primitive.ObjectID) (*model.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	if u, ok := f.users[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) List(_ context.Context, filter bson.M, page, pageSize int64) ([]*model.User, int64, error) {
	var out []*model.User
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, int64(len(out)), nil
}
func (f *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	_, ok := f.users[email]
	return ok, nil
}

func newTestUserSvc(repo repository.UserRepository) *UserService {
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpiresMinutes: 60}
	return NewUserService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)
}

func TestUserRegister(t *testing.T) {
	svc := newTestUserSvc(newFakeUserRepo())
	cases := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"注册学生成功", constants.RoleStudent, false},
		{"非法角色", "superadmin", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email := tc.name + "@example.com"
			_, err := svc.Register(context.Background(), "测试用户", email, "123456", tc.role)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserRegisterDuplicateEmail(t *testing.T) {
	svc := newTestUserSvc(newFakeUserRepo())
	_, err := svc.Register(context.Background(), "a", "dup@example.com", "123456", constants.RoleStudent)
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err = svc.Register(context.Background(), "b", "dup@example.com", "123456", constants.RoleStudent)
	if err == nil {
		t.Fatal("duplicate email should error")
	}
}

func TestUserLogin(t *testing.T) {
	svc := newTestUserSvc(newFakeUserRepo())
	_, err := svc.Register(context.Background(), "张三", "login@example.com", "123456", constants.RoleStudent)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	cases := []struct {
		name    string
		pwd     string
		wantErr bool
	}{
		{"正确密码", "123456", false},
		{"错误密码", "wrong-pass", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.Login(context.Background(), "login@example.com", tc.pwd)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Login() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
