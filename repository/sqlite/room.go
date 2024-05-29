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

func (rr *roomRepository) Create(ctx context.Context, room entity.Room) error {
	stmt, err := rr.db.Prepare("INSERT INTO rooms(id, name, private) values(?, ?, ?)")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = stmt.ExecContext(ctx, room.ID, room.Name, room.Private)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (rr *roomRepository) Get(ctx context.Context, name string) (*entity.Room, error) {
	var room entity.Room
	row := rr.db.QueryRowContext(ctx, "SELECT id, name, private FROM rooms WHERE name = ? LIMIT 1", name)

	if err := row.Scan(&room.ID, &room.Name, &room.Private); err != nil {
		log.Println(err)
		return nil, err
	}
	return &room, nil
}
