// Package config 集中解析环境变量配置。
package config

import (
	"github.com/caarlos0/env/v11"
)

// Config 应用配置，全部通过环境变量注入。
type Config struct {
	// 服务
	ServerPort         string `env:"SERVER_PORT" envDefault:"8080"`
	ServerHost         string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	RunMode            string `env:"GIN_MODE" envDefault:"release"`
	LogLevel           string `env:"LOG_LEVEL" envDefault:"info"`
	AllowedOrigins     string `env:"ALLOWED_ORIGINS" envDefault:"*"`
	RateLimitPerMinute int    `env:"RATE_LIMIT_PER_MINUTE" envDefault:"300"`

	// MongoDB
	MongoURI      string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDBName   string `env:"DB_NAME" envDefault:"onlineexam_db"`
	MongoUsername string `env:"DB_USER" envDefault:""`
	MongoPassword string `env:"DB_PASSWORD" envDefault:""`

	// Redis
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// JWT
	JWTSecret         string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpiresMinutes int    `env:"JWT_EXPIRES_MINUTES" envDefault:"720"`

	// 种子数据
	SeedEnabled bool `env:"SEED_ENABLED" envDefault:"true"`
}

// Load 从环境变量加载配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
