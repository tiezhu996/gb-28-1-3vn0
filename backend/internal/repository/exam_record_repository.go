package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	// SaveDraft 自动保存草稿：仅当记录仍为 in_progress 且请求版本号新于已存版本号时才写入（原子条件更新）。
	// 返回 ErrNotFound 表示记录不存在；ErrConflict 表示状态已变（已交卷等）或版本号过旧（乱序旧请求）。
	SaveDraft(ctx context.Context, id, studentID primitive.ObjectID, questions []model.AttemptQuestion, currentIndex int, version int64, savedAt time.Time) error
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

// SaveDraft 条件更新草稿：本人 + in_progress + 更新的版本号三者同时满足才写入，
// 杜绝乱序保存覆盖新答案，也不会写花已交卷的答卷。
func (r *MongoExamRecordRepository) SaveDraft(ctx context.Context, id, studentID primitive.ObjectID, questions []model.AttemptQuestion, currentIndex int, version int64, savedAt time.Time) error {
	filter := bson.M{
		"_id":          id,
		"student_id":   studentID,
		"status":       modelStatusInProgress(),
		"save_version": bson.M{"$lt": version},
	}
	update := bson.M{
		"$set": bson.M{
			"questions":     questions,
			"current_index": currentIndex,
			"save_version":  version,
			"last_saved_at": savedAt,
			"updated_at":    savedAt,
		},
	}
	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("save draft exam record: %w", err)
	}
	if res.MatchedCount == 0 {
		// 区分“记录不存在/非本人/已交卷”与“版本号过旧”，统一由 service 再查一次判定，这里先返回冲突。
		return ErrConflict
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
