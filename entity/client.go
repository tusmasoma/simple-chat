package entity

import "github.com/google/uuid"

type Client struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
