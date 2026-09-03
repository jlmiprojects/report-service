package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	HostPort string        `mapstructure:"host_port"`
	Password string        `mapstructure:"password"`
	Ssl      bool          `mapstructure:"ssl"`
	DB       int           `mapstructure:"db"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

func NewRedisClient(config *RedisConfig) (*RedisClient, error) {

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	rc := redis.NewClient(&redis.Options{
		ClientName: "redis-client",
		Addr:       config.HostPort,
		Password:   config.Password, // no password set
		DB:         config.DB,       // use default DB
	})

	status := rc.Ping(ctx)
	if status.Err() != nil {
		return nil, status.Err()
	}

	return &RedisClient{client: rc, config: config}, nil
}

func (r *RedisClient) SetEx(key string, value interface{}, ttl time.Duration) error {

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	status := r.client.SetEx(ctx, key, value, ttl)

	if status.Err() != nil {
		return status.Err()
	}

	return nil
}

func (r *RedisClient) GetString(key string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	status := r.client.Get(ctx, key)

	if status.Err() != nil {
		return "", status.Err()
	}

	return status.Val(), nil
}
