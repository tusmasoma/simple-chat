package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tusmasoma/simple-chat/config"
	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	hub        *Hub
	rooms      map[*Room]bool
	conn       *websocket.Conn
	send       chan []byte
	pubsubRepo repository.PubSubRepository
}

func NewClientWebSocketRepository(conn *websocket.Conn, hub *Hub, name string, id string, pubsubRepo repository.PubSubRepository) repository.ClientWebSocketRepository {
	return &Client{
		ID:         id,
		Name:       name,
		conn:       conn,
		hub:        hub,
		pubsubRepo: pubsubRepo,
	}
}

func (client *Client) ReadPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(config.MaxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(config.PongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(config.PongWait)); return nil })

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

func (client *Client) WritePump() {
	ticker := time.NewTicker(config.PingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
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
			client.conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
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
	var message entity.Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshalling JSON message: %v", err)
		return
	}

	// Attach the client object as the sender of the message.
	message.SenderID = client.ID

	switch message.Action {
	case config.SendMessageAction:
		roomID := message.TargetID
		if room := client.hub.findRoomByID(roomID); room != nil {
			room.broadcast <- &message
		}
	case config.JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case config.LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	case config.JoinRoomPrivateAction:
		client.handleJoinRoomMessage(message)
	}
}

func (client *Client) handleJoinRoomMessage(message entity.Message) {
	roomName := message.Content

	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(message entity.Message) {
	room := client.hub.findRoomByID(message.Content)
	if room == nil {
		return
	}
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}

	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message entity.Message) {
	target := client.hub.findClientByID(message.Content)
	if target == nil {
		return
	}

	roomName := message.Content + client.ID

	joinedRoom := client.joinRoom(roomName, target)
	if joinedRoom != nil {
		client.inviteTargetUser(target, joinedRoom)
	}
}

func (client *Client) joinRoom(roomName string, sender *Client) *Room {
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
	inviteMessage := &entity.Message{
		Action:   config.JoinRoomPrivateAction,
		Content:  target.ID,
		TargetID: room.ID,
		SenderID: client.ID,
	}

	if err := client.pubsubRepo.Publish(context.Background(), config.PubSubGeneralChannel, inviteMessage.Encode()); err != nil {
		log.Print(err)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender *Client) {
	message := entity.Message{
		Action:   config.RoomJoinedAction,
		TargetID: room.ID,
		SenderID: sender.ID,
	}

	client.send <- message.Encode()
}
