package config

import "time"

const (
	SendMessageAction     = "send_message"
	JoinRoomAction        = "join_room"
	LeaveRoomAction       = "leave_room"
	UserJoinedAction      = "user_joined"
	UserLeftAction        = "user_left"
	JoinRoomPrivateAction = "join_room_private"
	RoomJoinedAction      = "room-joined"
)

const (
	// Max wait time when writing a message to the peer.
	WriteWait = 10 * time.Second

	// Max wait time for the peer to read the next pong message.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Max message size allowed from peer.
	MaxMessageSize = 10000
)

const WelcomeMessage = "%s joined the room"
const GoodbyeMessage = "%s left the room"

const PubSubGeneralChannel = "general"
