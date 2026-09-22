// Package migrations 负责启动时建索引与种子数据。
package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes 为各集合创建唯一索引与常用查询索引。
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	indexes := []struct {
		coll string
		keys bson.D
		opts *options.IndexOptions
	}{
		{"users", bson.D{{Key: "email", Value: 1}}, options.Index().SetUnique(true)},
		{"users", bson.D{{Key: "role", Value: 1}}, nil},
		{"questions", bson.D{{Key: "subject", Value: 1}, {Key: "type", Value: 1}}, nil},
		{"questions", bson.D{{Key: "difficulty", Value: 1}}, nil},
		{"questions", bson.D{{Key: "knowledge_points", Value: 1}}, nil},
		{"exams", bson.D{{Key: "status", Value: 1}, {Key: "subject", Value: 1}}, nil},
		{"exam_records", bson.D{{Key: "exam_id", Value: 1}, {Key: "student_id", Value: 1}}, nil},
		{"exam_records", bson.D{{Key: "status", Value: 1}}, nil},
		{"wrong_books", bson.D{{Key: "student_id", Value: 1}, {Key: "question_id", Value: 1}}, options.Index().SetUnique(true)},
		{"audit_logs", bson.D{{Key: "created_at", Value: -1}}, nil},
		{"audit_logs", bson.D{{Key: "module", Value: 1}, {Key: "action", Value: 1}}, nil},
	}
	for _, idx := range indexes {
		_, err := db.Collection(idx.coll).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    idx.keys,
			Options: idx.opts,
		})
		if err != nil {
			return fmt.Errorf("create index %s: %w", idx.coll, err)
		}
	}
	return nil
}
