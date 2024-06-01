package entity

import (
	"encoding/json"
	"log"
)

type Message struct {
	Action   string `json:"action"`
	Content  string `json:"content"`
	TargetID string `json:"target"`
	SenderID string `json:"sender"`
}

func (message *Message) Encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	return json
}
