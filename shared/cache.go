package utils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	defaultTtl time.Duration
	client     *redis.Client
	config     *RedisConfig
}

func NewCache(config *RedisConfig, defaultTtl time.Duration) (*Cache, error) {

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	rc := redis.NewClient(&redis.Options{
		ClientName: "cache-client",
		Addr:       config.HostPort,
		Password:   config.Password, // no password set
		DB:         config.DB,       // use default DB
	})

	status := rc.Ping(ctx)
	if status.Err() != nil {
		return nil, status.Err()
	}

	c := Cache{
		defaultTtl: defaultTtl,
		client:     rc,
		config:     config,
	}

	return &c, nil

}

func (cache *Cache) Set(key string, value any) error {

	ctx, cancel := context.WithTimeout(context.Background(), cache.config.Timeout)
	defer cancel()

	j, err := json.Marshal(value)
	if err != nil {
		return err
	}
	status := cache.client.SetEx(ctx, key, j, cache.defaultTtl)

	if status.Err() != nil {
		return status.Err()
	}

	return nil

}

func (cache *Cache) Delete(key string) {

	ctx, cancel := context.WithTimeout(context.Background(), cache.config.Timeout)
	defer cancel()

	cache.client.Del(ctx, key)

}

func (cache *Cache) SetEx(key string, value string, ttl time.Duration) error {

	ctx, cancel := context.WithTimeout(context.Background(), cache.config.Timeout)
	defer cancel()

	j, err := json.Marshal(value)
	if err != nil {
		return err
	}
	status := cache.client.SetEx(ctx, key, j, ttl)

	if status.Err() != nil {
		return status.Err()
	}

	return nil

}

func (cache Cache) Get(key string, value any) error {

	ctx, cancel := context.WithTimeout(context.Background(), cache.config.Timeout)
	defer cancel()

	status := cache.client.Get(ctx, key)

	if status.Err() != nil {
		return status.Err()
	}

	b, err := status.Bytes()
	if err != nil {
		return status.Err()
	}

	err = json.Unmarshal(b, value)

	if err != nil {
		return err
	}

	return nil

}
