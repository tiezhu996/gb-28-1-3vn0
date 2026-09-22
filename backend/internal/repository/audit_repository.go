package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onlineexam/onlineexam/internal/model"
)

// AuditRepository 审计日志仓储接口。
type AuditRepository interface {
	Create(ctx context.Context, a *model.AuditLog) error
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.AuditLog, int64, error)
}

// MongoAuditRepository MongoDB 审计日志仓储实现。
type MongoAuditRepository struct {
	coll *mongo.Collection
}

// NewMongoAuditRepository 构造审计日志仓储。
func NewMongoAuditRepository(db *mongo.Database) *MongoAuditRepository {
	return &MongoAuditRepository{coll: db.Collection("audit_logs")}
}

func (r *MongoAuditRepository) Create(ctx context.Context, a *model.AuditLog) error {
	_, err := r.coll.InsertOne(ctx, a)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *MongoAuditRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.AuditLog, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var list []*model.AuditLog
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode audit logs: %w", err)
	}
	return list, total, nil
}
