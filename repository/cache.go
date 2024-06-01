package repository

import "context"

type CacheRepository interface {
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string) (any, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
}
