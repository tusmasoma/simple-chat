package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/tusmasoma/simple-chat/repository"
)

type userCacheRepository struct {
	client *redis.Client
}

func NewUserRepository(client *redis.Client) repository.UserCacheRepository {
	return &userCacheRepository{
		client: client,
	}
}

func (ur *userCacheRepository) GetUserSession(ctx context.Context, userID string) (string, error) {
	return ur.client.Get(ctx, userID).Result()
}

func (ur *userCacheRepository) SetUserSession(ctx context.Context, userID string, sessionData string) error {
	return ur.client.Set(ctx, userID, sessionData, 0).Err()
}
