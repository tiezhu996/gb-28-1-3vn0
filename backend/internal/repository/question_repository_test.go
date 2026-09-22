package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onlineexam/onlineexam/internal/model"
)

// TestQuestionRepositoryMongo 集成测试：需要 MONGO_TEST_URI。
func TestQuestionRepositoryMongo(t *testing.T) {
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
	repo := NewMongoQuestionRepository(db)

	q := &model.Question{
		ID: primitive.NewObjectID(), Type: "single", Subject: "语文",
		KnowledgePoints: []string{"古诗"}, Difficulty: "easy", Content: "床前明月光？",
		Answer: "A", Score: 5, Status: "published", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repo.Create(ctx, q); err != nil {
		t.Fatalf("create: %v", err)
	}
	list, total, err := repo.List(ctx, bson.M{"subject": "语文"}, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("list empty")
	}
	if _, err := repo.RandomPick(ctx, bson.M{"subject": "语文"}, 1); err != nil {
		t.Fatalf("random pick: %v", err)
	}
	if err := repo.Delete(ctx, q.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_ = db.Drop(ctx)
}
