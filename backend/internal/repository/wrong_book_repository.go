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

// WrongBookRepository 错题本仓储接口。
type WrongBookRepository interface {
	Create(ctx context.Context, w *model.WrongBook) error
	Update(ctx context.Context, w *model.WrongBook) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.WrongBook, error)
	FindByStudentAndQuestion(ctx context.Context, studentID, questionID primitive.ObjectID) (*model.WrongBook, error)
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.WrongBook, int64, error)
}

// MongoWrongBookRepository MongoDB 错题本仓储实现。
type MongoWrongBookRepository struct {
	coll *mongo.Collection
}

// NewMongoWrongBookRepository 构造错题本仓储。
func NewMongoWrongBookRepository(db *mongo.Database) *MongoWrongBookRepository {
	return &MongoWrongBookRepository{coll: db.Collection("wrong_books")}
}

func (r *MongoWrongBookRepository) Create(ctx context.Context, w *model.WrongBook) error {
	_, err := r.coll.InsertOne(ctx, w)
	if err != nil {
		return fmt.Errorf("create wrong book: %w", err)
	}
	return nil
}

func (r *MongoWrongBookRepository) Update(ctx context.Context, w *model.WrongBook) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": w.ID}, w)
	if err != nil {
		return fmt.Errorf("update wrong book: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update wrong book: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoWrongBookRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete wrong book: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("delete wrong book: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoWrongBookRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.WrongBook, error) {
	var w model.WrongBook
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&w); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find wrong book by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find wrong book by id: %w", err)
	}
	return &w, nil
}

func (r *MongoWrongBookRepository) FindByStudentAndQuestion(ctx context.Context, studentID, questionID primitive.ObjectID) (*model.WrongBook, error) {
	var w model.WrongBook
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID, "question_id": questionID}).Decode(&w)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find wrong book by student and question: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find wrong book by student and question: %w", err)
	}
	return &w, nil
}

func (r *MongoWrongBookRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.WrongBook, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count wrong books: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list wrong books: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var list []*model.WrongBook
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode wrong books: %w", err)
	}
	return list, total, nil
}
