package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/tusmasoma/simple-chat/repository"
)

type pubsubRepository struct {
	client *redis.Client
}

func NewPubSubRepository(client *redis.Client) repository.PubSubRepository {
	return &pubsubRepository{
		client,
	}
}

func (r *pubsubRepository) Publish(ctx context.Context, channel string, message any) error {
	return r.client.Publish(ctx, channel, message).Err()
}

func (r *pubsubRepository) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}
