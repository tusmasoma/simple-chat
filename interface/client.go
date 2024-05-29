package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

const (
	// Max wait time when writing a message to the peer.
	writeWait = 10 * time.Second

	// Max wait time for the peer to read the next pong message.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Max message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	ID         uuid.UUID `json:"id"`
	conn       *websocket.Conn
	hub        *Hub
	send       chan []byte
	rooms      map[*Room]bool
	Name       string `json:"name"`
	pubsubRepo repository.PubSubRepository
}

func newClient(conn *websocket.Conn, hub *Hub, name string, pubsubRepo repository.PubSubRepository) *Client {
	return &Client{
		ID:         uuid.New(),
		conn:       conn,
		hub:        hub,
		send:       make(chan []byte, 256),
		rooms:      make(map[*Room]bool),
		Name:       name,
		pubsubRepo: pubsubRepo,
	}
}

func (c *Client) ToUser() *entity.User {
	return &entity.User{
		ID:   c.ID,
		Name: c.Name,
	}
}

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	client.hub.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
	close(client.send)
	client.conn.Close()
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshalling JSON message: %v", err)
		return
	}

	// Attach the client object as the sender of the message.
	message.Sender = client.ToUser()

	switch message.Action {
	case SendMessageAction:
		roomID := message.Target.ID.String()
		if room := client.hub.findRoomByID(roomID); room != nil {
			room.broadcast <- &message
		}
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	case JoinRoomPrivateAction:
		client.handleJoinRoomMessage(message)
	}
}

func (client *Client) handleJoinRoomMessage(message Message) {
	roomName := message.Message

	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	room := client.hub.findRoomByID(message.Message)
	if room == nil {
		return
	}
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}

	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message Message) {
	target := client.hub.findClientByID(message.Message)
	if target == nil {
		return
	}

	roomName := message.Message + client.ID.String()

	joinedRoom := client.joinRoom(roomName, target.ToUser())
	if joinedRoom != nil {
		client.inviteTargetUser(target, joinedRoom)
	}
}

func (client *Client) joinRoom(roomName string, sender *entity.User) *Room {
	room := client.hub.findRoomByName(roomName)
	if room == nil {
		room = client.hub.createRoom(roomName, sender != nil)
	}

	if sender == nil && room.Private {
		return nil
	}

	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.register <- client
		client.notifyRoomJoined(room, sender)
	}
	return room
}

// Send out invite message over pub/sub in the general channel
func (client *Client) inviteTargetUser(target *Client, room *Room) {
	inviteMessage := &Message{
		Action:  JoinRoomPrivateAction,
		Message: target.ID.String(),
		Target:  room,
		Sender:  client.ToUser(),
	}

	if err := client.pubsubRepo.Publish(context.Background(), PubSubGeneralChannel, inviteMessage.encode()); err != nil {
		log.Print(err)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender *entity.User) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}

	client.send <- message.encode()
}
