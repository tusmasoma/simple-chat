package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type ClientRepository interface {
	Create(ctx context.Context, client entity.Client) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*entity.Client, error)
	List(ctx context.Context) ([]*entity.Client, error)
}
