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

func (cr *clientRepository) Create(ctx context.Context, client entity.Client) {
	stmt, err := cr.db.Prepare("INSERT INTO clients(id, name) values(?, ?)")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.ExecContext(ctx, client.ID, client.Name)
	if err != nil {
		log.Println(err)
	}
}

func (cr *clientRepository) Delete(ctx context.Context, client entity.Client) {
	stmt, err := cr.db.Prepare("DELETE FROM clients WHERE id = ?")
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.ExecContext(ctx, client.ID)
	if err != nil {
		log.Println(err)
	}
}

func (cr *clientRepository) Get(ctx context.Context, ID string) *entity.Client {
	var client entity.Client
	row := cr.db.QueryRowContext(ctx, "SELECT id, name FROM clients WHERE id = ? LIMIT 1", ID)

	if err := row.Scan(&client.ID, &client.Name); err != nil {
		log.Println(err)
		return nil
	}
	return &client
}

func (cr *clientRepository) List(ctx context.Context) []*entity.Client {
	var clients []*entity.Client
	rows, err := cr.db.QueryContext(ctx, "SELECT id, name FROM clients")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var client entity.Client
		if err := rows.Scan(&client.ID, &client.Name); err != nil {
			log.Println(err)
			return nil
		}
		clients = append(clients, &client)
	}
	return clients
}
