package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type PubSubRepository interface {
	Publish(ctx context.Context, channel string, message any) error
	Subscribe(ctx context.Context, channel string) *redis.PubSub // TODO: *redis.PubSub を汎用化する
}
