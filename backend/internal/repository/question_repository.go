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

// QuestionRepository 题库仓储接口。
type QuestionRepository interface {
	Create(ctx context.Context, q *model.Question) error
	CreateMany(ctx context.Context, qs []*model.Question) error
	Update(ctx context.Context, q *model.Question) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Question, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.Question, error)
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Question, int64, error)
	RandomPick(ctx context.Context, filter bson.M, limit int) ([]*model.Question, error)
}

// MongoQuestionRepository MongoDB 题库仓储实现。
type MongoQuestionRepository struct {
	coll *mongo.Collection
}

// NewMongoQuestionRepository 构造题库仓储。
func NewMongoQuestionRepository(db *mongo.Database) *MongoQuestionRepository {
	return &MongoQuestionRepository{coll: db.Collection("questions")}
}

func (r *MongoQuestionRepository) Create(ctx context.Context, q *model.Question) error {
	_, err := r.coll.InsertOne(ctx, q)
	if err != nil {
		return fmt.Errorf("create question: %w", err)
	}
	return nil
}

func (r *MongoQuestionRepository) CreateMany(ctx context.Context, qs []*model.Question) error {
	docs := make([]interface{}, 0, len(qs))
	for _, q := range qs {
		docs = append(docs, q)
	}
	_, err := r.coll.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("create many questions: %w", err)
	}
	return nil
}

func (r *MongoQuestionRepository) Update(ctx context.Context, q *model.Question) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": q.ID}, q)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update question: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoQuestionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("delete question: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoQuestionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Question, error) {
	var q model.Question
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&q); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find question by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find question by id: %w", err)
	}
	return &q, nil
}

func (r *MongoQuestionRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.Question, error) {
	if len(ids) == 0 {
		return []*model.Question{}, nil
	}
	cur, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("find questions by ids: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var qs []*model.Question
	if err := cur.All(ctx, &qs); err != nil {
		return nil, fmt.Errorf("decode questions: %w", err)
	}
	return qs, nil
}

func (r *MongoQuestionRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.Question, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count questions: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list questions: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var qs []*model.Question
	if err := cur.All(ctx, &qs); err != nil {
		return nil, 0, fmt.Errorf("decode questions: %w", err)
	}
	return qs, total, nil
}

// RandomPick 使用 $sample 随机抽取题目（自动组卷、随机题序）。
func (r *MongoQuestionRepository) RandomPick(ctx context.Context, filter bson.M, limit int) ([]*model.Question, error) {
	if limit <= 0 {
		return []*model.Question{}, nil
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.M{"size": limit}}},
	}
	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("random pick questions: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var qs []*model.Question
	if err := cur.All(ctx, &qs); err != nil {
		return nil, fmt.Errorf("decode random questions: %w", err)
	}
	return qs, nil
}
