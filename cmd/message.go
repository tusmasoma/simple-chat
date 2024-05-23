package main

import (
	"encoding/json"
	"log"
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
	Action  string  `json:"action"`
	Message string  `json:"message"`
	Target  *Room   `json:"target"`
	Sender  *Client `json:"sender"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	return json
}