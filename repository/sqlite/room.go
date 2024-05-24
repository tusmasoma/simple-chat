package sqlite

import (
	"context"
	"database/sql"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

type roomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) repository.RoomRepository {
	return &roomRepository{
		db,
	}
}

func (repo *roomRepository) AddRoom(ctx context.Context, room entity.Room) {
	stmt, err := repo.db.Prepare("INSERT INTO room(id, name, private) values(?, ?, ?)")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.ExecContext(ctx, room.ID, room.Name, room.Private)
	if err != nil {
		log.Println(err)
	}
}

func (repo *roomRepository) FindRoomByName(ctx context.Context, name string) *entity.Room {
	var room entity.Room
	row := repo.db.QueryRowContext(ctx, "SELECT id, name, private FROM room WHERE name = ? LIMIT 1", name)

	if err := row.Scan(&room.ID, &room.Name, &room.Private); err != nil {
		log.Println(err)
		return nil
	}
	return &room
}
