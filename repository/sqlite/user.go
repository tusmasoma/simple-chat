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

func (repo *userRepository) AddUser(ctx context.Context, user entity.User) {
	stmt, err := repo.db.Prepare("INSERT INTO user(id, name) values(?, ?)")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.ExecContext(ctx, user.ID, user.Name)
	if err != nil {
		log.Println(err)
	}
}

func (repo *userRepository) RemoveUser(ctx context.Context, user entity.User) {
	stmt, err := repo.db.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.ExecContext(ctx, user.ID)
	if err != nil {
		log.Println(err)
	}
}

func (repo *userRepository) FindUserById(ctx context.Context, ID string) *entity.User {
	var user entity.User
	row := repo.db.QueryRowContext(ctx, "SELECT id, name FROM user WHERE id = ? LIMIT 1", ID)

	if err := row.Scan(&user.ID, &user.Name); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func (repo *userRepository) GetAllUsers(ctx context.Context) []entity.User {
	var users []entity.User
	rows, err := repo.db.QueryContext(ctx, "SELECT id, name FROM user")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name); err != nil {
			log.Println(err)
			return nil
		}
		users = append(users, user)
	}
	return users
}
