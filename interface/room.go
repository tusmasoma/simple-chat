package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/tusmasoma/simple-chat/repository"
)

const welcomeMessage = "%s joined the room"
const goodbyeMessage = "%s left the room"

type Room struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	pubsubRepo repository.PubSubRepository
}

// NewRoom creates a new Room
func NewRoom(name string, private bool, pubsubRepo repository.PubSubRepository) *Room {
	return &Room{
		ID:         uuid.New(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		pubsubRepo: pubsubRepo,
	}
}

// Run starts the room and listens for incoming messages
func (room *Room) Run() {
	ctx := context.Background()
	go room.subscribeToRoomMessages(ctx)

	for {
		select {

		case client := <-room.register:
			room.registerClientInRoom(client)

		case client := <-room.unregister:
			room.unregisterClientInRoom(client)

		case message := <-room.broadcast:
			room.publishRoomMessage(ctx, message.encode())
		}
	}
}

func (room *Room) registerClientInRoom(client *Client) {
	if !room.Private {
		room.notifyClientJoined(client)
	}
	room.clients[client] = true
}

func (room *Room) unregisterClientInRoom(client *Client) {
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action:  SendMessageAction,
		Message: fmt.Sprintf(welcomeMessage, client.Name),
		Target:  room,
		Sender:  user,
	}

	room.broadcastToClientsInRoom(message.encode())
}

// TODO: ここでは、チャンネルをroomの名前にしている。一意せいないので命名考える
func (room *Room) publishRoomMessage(ctx context.Context, message []byte) {
	err := room.pubsubRepo.Publish(ctx, room.Name, message)
	if err != nil {
		log.Print(err)
	}
}

func (room *Room) subscribeToRoomMessages(ctx context.Context) {
	pubsub := room.pubsubRepo.Subscribe(ctx, room.Name)

	ch := pubsub.Channel()

	for msg := range ch {
		room.broadcastToClientsInRoom([]byte(msg.Payload))
	}
}
