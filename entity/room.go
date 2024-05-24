package entity

import "github.com/google/uuid"

type Room struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Private bool      `json:"private"`
}
