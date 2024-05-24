package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type RoomRepository interface {
	AddRoom(ctx context.Context, room entity.Room)
	FindRoomByName(ctx context.Context, name string) *entity.Room
}
