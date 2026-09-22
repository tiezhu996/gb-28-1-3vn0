package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onlineexam/onlineexam/internal/model"
)

// ExamRepository 试卷仓储接口。
type ExamRepository interface {
	Create(ctx context.Context, e *model.Exam) error
	Update(ctx context.Context, e *model.Exam) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Exam, error)
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Exam, int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
}

// MongoExamRepository MongoDB 试卷仓储实现。
type MongoExamRepository struct {
	coll *mongo.Collection
}

// NewMongoExamRepository 构造试卷仓储。
func NewMongoExamRepository(db *mongo.Database) *MongoExamRepository {
	return &MongoExamRepository{coll: db.Collection("exams")}
}

func (r *MongoExamRepository) Create(ctx context.Context, e *model.Exam) error {
	_, err := r.coll.InsertOne(ctx, e)
	if err != nil {
		return fmt.Errorf("create exam: %w", err)
	}
	return nil
}

func (r *MongoExamRepository) Update(ctx context.Context, e *model.Exam) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": e.ID}, e)
	if err != nil {
		return fmt.Errorf("update exam: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update exam: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoExamRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete exam: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("delete exam: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoExamRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Exam, error) {
	var e model.Exam
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&e); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find exam by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find exam by id: %w", err)
	}
	return &e, nil
}

func (r *MongoExamRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Exam, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count exams: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list exams: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var exams []*model.Exam
	if err := cur.All(ctx, &exams); err != nil {
		return nil, 0, fmt.Errorf("decode exams: %w", err)
	}
	return exams, total, nil
}

func (r *MongoExamRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"status": status})
	if err != nil {
		return 0, fmt.Errorf("count exams by status: %w", err)
	}
	return count, nil
}
