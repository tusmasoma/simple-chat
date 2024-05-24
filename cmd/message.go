package main

import (
	"encoding/json"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
)

const (
	SendMessageAction     = "send_message"
	JoinRoomAction        = "join_room"
	LeaveRoomAction       = "leave_room"
	UserJoinedAction      = "user_joined"
	UserLeftAction        = "user_left"
	JoinRoomPrivateAction = "join_room_private"
	RoomJoinedAction      = "room-joined"
)

type Message struct {
	Action  string       `json:"action"`
	Message string       `json:"message"`
	Target  *Room        `json:"target"`
	Sender  *entity.User `json:"sender"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	return json
}

// UnmarshalJSON custom unmarshel to create a Client instance for Sender
func (message *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	msg := &struct {
		Sender Client `json:"sender"`
		*Alias
	}{
		Alias: (*Alias)(message),
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	message.Sender = &msg.Sender
	return nil
}
