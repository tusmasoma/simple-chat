package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type RoomRepository interface {
	Create(ctx context.Context, room entity.Room) error
	Get(ctx context.Context, name string) *entity.Room // TODO: Change to ID
}
