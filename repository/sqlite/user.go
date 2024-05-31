package sqlite

import (
	"context"
	"database/sql"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{
		db,
	}
}

func (ur *userRepository) Create(ctx context.Context, client entity.User) error {
	stmt, err := ur.db.Prepare("INSERT INTO users(id, name) values(?, ?)")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = stmt.ExecContext(ctx, client.ID, client.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (ur *userRepository) Delete(ctx context.Context, id string) error {
	stmt, err := ur.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (ur *userRepository) Get(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	row := ur.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ? LIMIT 1", id)

	if err := row.Scan(&user.ID, &user.Name); err != nil {
		log.Println(err)
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) List(ctx context.Context) ([]*entity.User, error) {
	var users []*entity.User
	rows, err := ur.db.QueryContext(ctx, "SELECT id, name FROM users")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			log.Println(err)
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
