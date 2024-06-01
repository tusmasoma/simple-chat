package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/tusmasoma/simple-chat/repository"
)

var ErrCacheMiss = errors.New("cache: key not found")

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) repository.CacheRepository {
	return &redisRepository{
		client: client,
	}
}

func (rr *redisRepository) Set(ctx context.Context, key string, value any) error {
	err := rr.client.Set(ctx, key, value, 0).Err()
	return err
}

func (rr *redisRepository) Get(ctx context.Context, key string) (any, error) {
	val, err := rr.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	} else if err != nil {
		return nil, err
	}
	return val, nil
}

func (rr *redisRepository) Delete(ctx context.Context, key string) error {
	err := rr.client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	} else if err != nil {
		return err
	}
	return nil
}

func (rr *redisRepository) Exists(ctx context.Context, key string) bool {
	val := rr.client.Exists(ctx, key).Val()
	return val > 0
}
