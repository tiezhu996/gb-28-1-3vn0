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

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.User, int64, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// MongoUserRepository MongoDB 用户仓储实现。
type MongoUserRepository struct {
	coll *mongo.Collection
}

// NewMongoUserRepository 构造用户仓储。
func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{coll: db.Collection("users")}
}

func (r *MongoUserRepository) Create(ctx context.Context, u *model.User) error {
	_, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("create user: %w", ErrConflict)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *MongoUserRepository) Update(ctx context.Context, u *model.User) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": u.ID}, u)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("update user: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("delete user: %w", ErrNotFound)
	}
	return nil
}

func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find user by id: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("find user by email: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *MongoUserRepository) List(ctx context.Context, filter bson.M, page, pageSize int64) ([]*model.User, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()
	var users []*model.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, 0, fmt.Errorf("decode users: %w", err)
	}
	return users, total, nil
}

func (r *MongoUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("exists user by email: %w", err)
	}
	return count > 0, nil
}
