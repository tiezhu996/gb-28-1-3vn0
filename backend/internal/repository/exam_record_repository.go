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

// ExamRecordRepository 考试记录仓储接口。
type ExamRecordRepository interface {
	Create(ctx context.Context, r *model.ExamRecord) error
	Update(ctx context.Context, r *model.ExamRecord) error
	// SaveDraft 乐观锁保存草稿：仅当答卷仍属于本人、进行中且 answer_version 与期望版本一致时替换。
	// 版本不一致（乱序的旧请求）返回 ErrVersionConflict。
	SaveDraft(ctx context.Context, r *model.ExamRecord, expectedVersion int64) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.ExamRecord, error)
	FindActiveByExamAndStudent(ctx context.Context, examID, studentID primitive.ObjectID) (*model.ExamRecord, error)
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.ExamRecord, int64, error)
	ListAll(ctx context.Context, filter bson.M) ([]*model.ExamRecord, error)
	CountByExamAndStatus(ctx context.Context, examID primitive.ObjectID, statuses []string) (int64, error)
}

// MongoExamRecordRepository MongoDB 考试记录仓储实现。
type MongoExamRecordRepository struct {
	coll *mongo.Collection
}

// NewMongoExamRecordRepository 构造考试记录仓储。
func NewMongoExamRecordRepository(db *mongo.Database) *MongoExamRecordRepository {
	return &MongoExamRecordRepository{coll: db.Collection("exam_records")}
}

func (r *MongoExamRecordRepository) Create(ctx context.Context, rec *model.ExamRecord) error {
	_, err := r.coll.InsertOne(ctx, rec)
	if err != nil {
		return fmt.Errorf("create exam record: %w", err)
	}
	return nil
}

func (r *MongoExamRecordRepository) Update(ctx context.Context, rec *model.ExamRecord) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": rec.ID}, rec)
	if err != nil {
		return fmt.Errorf("update exam record: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update exam record: %w", ErrNotFound)
	}
	return nil
}

// SaveDraft 条件替换整份答卷：过滤条件同时校验归属人、进行中状态与草稿版本，
// 保证“他人答卷 / 已交卷 / 旧版本请求”都不会覆盖最新答案（原子操作，避免读-改-写竞态）。
func (r *MongoExamRecordRepository) SaveDraft(ctx context.Context, rec *model.ExamRecord, expectedVersion int64) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{
		"_id":            rec.ID,
		"student_id":     rec.StudentID,
		"status":         modelStatusInProgress(),
		"answer_version": expectedVersion,
	}, rec)
	if err != nil {
		return fmt.Errorf("save draft exam record: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("save draft exam record: %w", ErrVersionConflict)
	}
	return nil
}

func (r *MongoExamRecordRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&rec); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find exam record by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find exam record by id: %w", err)
	}
	return &rec, nil
}

func (r *MongoExamRecordRepository) FindActiveByExamAndStudent(ctx context.Context, examID, studentID primitive.ObjectID) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	err := r.coll.FindOne(ctx, bson.M{
		"exam_id":    examID,
		"student_id": studentID,
		"status":     modelStatusInProgress(),
	}).Decode(&rec)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find active exam record: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find active exam record: %w", err)
	}
	return &rec, nil
}

func (r *MongoExamRecordRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.ExamRecord, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count exam records: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"started_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list exam records: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var recs []*model.ExamRecord
	if err := cur.All(ctx, &recs); err != nil {
		return nil, 0, fmt.Errorf("decode exam records: %w", err)
	}
	return recs, total, nil
}

func (r *MongoExamRecordRepository) ListAll(ctx context.Context, filter bson.M) ([]*model.ExamRecord, error) {
	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list all exam records: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var recs []*model.ExamRecord
	if err := cur.All(ctx, &recs); err != nil {
		return nil, fmt.Errorf("decode all exam records: %w", err)
	}
	return recs, nil
}

func (r *MongoExamRecordRepository) CountByExamAndStatus(ctx context.Context, examID primitive.ObjectID, statuses []string) (int64, error) {
	filter := bson.M{"exam_id": examID}
	if len(statuses) > 0 {
		filter["status"] = bson.M{"$in": statuses}
	}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count exam records by status: %w", err)
	}
	return count, nil
}

// modelStatusInProgress 避免 repository 直接依赖 constants 的循环，这里硬编码状态值。
// 注意：该状态枚举同时在 constants/enums.go、service 状态机、formatters、前端 constants 中出现。
func modelStatusInProgress() string {
	return "in_progress"
}
