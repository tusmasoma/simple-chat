package repository

import (
	"context"

	"github.com/tusmasoma/simple-chat/entity"
)

type UserRepository interface {
	AddUser(ctx context.Context, user entity.User)
	RemoveUser(ctx context.Context, user entity.User)
	FindUserById(ctx context.Context, ID string) *entity.User
	GetAllUsers(ctx context.Context) []*entity.User
}
