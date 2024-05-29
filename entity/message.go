package entity

import (
	"encoding/json"
	"log"
)

type Message struct {
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	ClientID string `json:"client_id"`
	Content  string `json:"content"`
	Action   string `json:"action"`
}

func (message *Message) Encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	return json
}
