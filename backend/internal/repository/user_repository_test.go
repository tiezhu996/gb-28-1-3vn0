package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onlineexam/onlineexam/internal/model"
)

// TestUserRepositoryMongo 集成测试：需要 MONGO_TEST_URI 环境变量（本地/CI 未配置时跳过）。
func TestUserRepositoryMongo(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI 未设置，跳过 MongoDB 集成测试")
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = client.Disconnect(ctx) }()
	db := client.Database("onlineexam_test")
	repo := NewMongoUserRepository(db)

	email := "test-" + primitive.NewObjectID().Hex() + "@example.com"
	u := &model.User{
		ID:           primitive.NewObjectID(),
		Name:         "测试",
		Email:        email,
		PasswordHash: "hash",
		Role:         "student",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if got.Name != "测试" {
		t.Fatalf("name mismatch: %s", got.Name)
	}
	if err := repo.Delete(ctx, u.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_ = db.Drop(ctx)
}
