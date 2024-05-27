package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type ClientRepository interface {
	Create(ctx context.Context, client entity.Client)
	Delete(ctx context.Context, client entity.Client)
	Get(ctx context.Context, ID string) *entity.Client
	List(ctx context.Context) []*entity.Client
}
