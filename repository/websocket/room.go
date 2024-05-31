package websocket

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/tusmasoma/simple-chat/config"
	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

type Room struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *entity.Message
	Private    bool `json:"private"`
	pubsubRepo repository.PubSubRepository
}

func NewRoom(name string, private bool, pubsub repository.PubSubRepository) *Room {
	return &Room{
		ID:         uuid.New().String(),
		Name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *entity.Message),
		Private:    private,
		pubsubRepo: pubsub,
	}
}

// NewRoom creates a new Room
func NewRoomWebSocketRepository(name string, private bool, pubsubRepo repository.PubSubRepository) repository.RoomWebSocketRepository {
	return &Room{
		ID:         uuid.New().String(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *entity.Message),
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
			room.publishRoomMessage(ctx, message.Encode())
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
	message := &entity.Message{
		Action:   config.SendMessageAction,
		Content:  fmt.Sprintf(config.WelcomeMessage, client.Name),
		TargetID: room.ID,
		SenderID: client.ID,
	}

	room.broadcastToClientsInRoom(message.Encode())
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
