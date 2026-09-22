// Package database 管理 MongoDB 与 Redis 连接。
package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/onlineexam/onlineexam/internal/config"
)

// Database 聚合 MongoDB 客户端与 Redis 客户端。
type Database struct {
	Mongo *mongo.Client
	DB    *mongo.Database
	Redis *redis.Client
}

// Connect 建立 MongoDB 与 Redis 连接。
func Connect(ctx context.Context, cfg *config.Config) (*Database, error) {
	mongoCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	uri := cfg.MongoURI
	if cfg.MongoUsername != "" {
		host := strings.TrimPrefix(cfg.MongoURI, "mongodb://")
		uri = fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=%s",
			cfg.MongoUsername, cfg.MongoPassword, host, cfg.MongoDBName, cfg.MongoDBName)
	}
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}
	if err := client.Ping(mongoCtx, nil); err != nil {
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Database{
		Mongo: client,
		DB:    client.Database(cfg.MongoDBName),
		Redis: rdb,
	}, nil
}

// Close 关闭所有连接。
func (d *Database) Close(ctx context.Context) error {
	if d.Mongo != nil {
		if err := d.Mongo.Disconnect(ctx); err != nil {
			return err
		}
	}
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			return err
		}
	}
	return nil
}
