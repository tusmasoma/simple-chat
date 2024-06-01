package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type UserRepository interface {
	Create(ctx context.Context, client entity.User) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
}

type UserCacheRepository interface {
	SetUserSession(ctx context.Context, userID string, sessionData string) error
	GetUserSession(ctx context.Context, userID string) (string, error)
}
