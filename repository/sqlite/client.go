package sqlite

import (
	"context"
	"database/sql"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

type clientRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.ClientRepository {
	return &clientRepository{
		db,
	}
}

func (cr *clientRepository) Create(ctx context.Context, client entity.Client) error {
	stmt, err := cr.db.Prepare("INSERT INTO clients(id, name) values(?, ?)")
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

func (cr *clientRepository) Delete(ctx context.Context, id string) error {
	stmt, err := cr.db.Prepare("DELETE FROM clients WHERE id = ?")
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

func (cr *clientRepository) Get(ctx context.Context, id string) (*entity.Client, error) {
	var client entity.Client
	row := cr.db.QueryRowContext(ctx, "SELECT id, name FROM clients WHERE id = ? LIMIT 1", id)

	if err := row.Scan(&client.ID, &client.Name); err != nil {
		log.Println(err)
		return nil, err
	}
	return &client, nil
}

func (cr *clientRepository) List(ctx context.Context) ([]*entity.Client, error) {
	var clients []*entity.Client
	rows, err := cr.db.QueryContext(ctx, "SELECT id, name FROM clients")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var client entity.Client
		if err := rows.Scan(&client.ID, &client.Name); err != nil {
			log.Println(err)
			return nil, err
		}
		clients = append(clients, &client)
	}
	return clients, nil
}
