package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tusmasoma/simple-chat/config"
	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
	roomRepo   repository.RoomRepository
	userRepo   repository.UserRepository
	pubsubRepo repository.PubSubRepository
	users      []*entity.User
}

// NewWebsocketServer creates a new WsServer type
func NewHubWebSocketRepository(ctx context.Context, roomRepo repository.RoomRepository, userRepo repository.UserRepository, pubsubRepo repository.PubSubRepository) repository.HubWebSocketRepository {
	hub := &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
		roomRepo:   roomRepo,
		userRepo:   userRepo,
		pubsubRepo: pubsubRepo,
	}

	hub.users, _ = userRepo.List(ctx)

	return hub
}

// Run starts the server and listens for incoming messages
func (h *Hub) Run() {
	go h.listenPubSubChannel()

	for {
		select {

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToClients(message)
		}

	}
}

func (h *Hub) registerClient(client *Client) {
	h.userRepo.Create(
		context.Background(),
		entity.User{
			ID:   client.ID,
			Name: client.Name,
		},
	)

	h.publishClientJoined(context.Background(), client)

	h.listOnlineClients(client)
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)

		h.publishClientLeft(context.Background(), client)
	}
}

func (h *Hub) broadcastToClients(message []byte) {
	for client := range h.clients {
		client.send <- message
	}
}

func (h *Hub) publishClientJoined(ctx context.Context, client *Client) error {
	if err := h.userRepo.Create(ctx, entity.User{
		ID:   client.ID,
		Name: client.Name,
	}); err != nil {
		log.Println(err)
		return err
	}

	message := &entity.Message{
		Action:   config.UserJoinedAction,
		SenderID: client.ID,
	}
	if err := h.pubsubRepo.Publish(ctx, config.PubSubGeneralChannel, message.Encode()); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// PublishClientLeft publishes a message to the general channel when a client leaves the server
func (h *Hub) publishClientLeft(ctx context.Context, client *Client) error {
	if err := h.userRepo.Delete(ctx, client.ID); err != nil {
		log.Println(err)
		return err
	}

	message := &entity.Message{
		Action:   config.UserLeftAction,
		SenderID: client.ID,
	}
	if err := h.pubsubRepo.Publish(ctx, config.PubSubGeneralChannel, message.Encode()); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (h *Hub) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range h.rooms {
		if room.Name == name {
			foundRoom = room
			break
		}
	}

	if foundRoom == nil {
		foundRoom = h.runRoomFromRepository(name)
	}

	return foundRoom
}

func (h *Hub) runRoomFromRepository(name string) *Room {
	var room *Room
	roomEntity, _ := h.roomRepo.Get(context.Background(), name)
	if roomEntity != nil {
		room = NewRoom(roomEntity.Name, roomEntity.Private, h.pubsubRepo)
		room.ID = roomEntity.ID

		go room.Run()
		h.rooms[room] = true
	}
	return room
}

func (h *Hub) findRoomByID(id string) *Room {
	for room := range h.rooms {
		if room.ID == id {
			return room
		}
	}
	return nil
}

func (h *Hub) findClientByID(id string) *Client {
	for client := range h.clients {
		if client.ID == id {
			return client
		}
	}
	return nil
}

func (h *Hub) createRoom(name string, private bool) *Room {
	room := NewRoom(name, private, h.pubsubRepo)

	h.roomRepo.Create(context.Background(), entity.Room{
		ID:      room.ID,
		Name:    room.Name,
		Private: room.Private,
	})

	go room.Run()
	h.rooms[room] = true
	return room
}

func (h *Hub) notifyClientJoined(client *Client) {
	message := &entity.Message{
		Action:   config.UserJoinedAction,
		Content:  fmt.Sprintf(config.WelcomeMessage, client.Name),
		SenderID: client.ID,
	}
	h.broadcastToClients(message.Encode())
}

func (h *Hub) notifyClientLeft(client *Client) {
	message := &entity.Message{
		Action:   config.UserLeftAction,
		Content:  fmt.Sprintf(config.GoodbyeMessage, client.Name),
		SenderID: client.ID,
	}
	h.broadcastToClients(message.Encode())
}

func (h *Hub) listOnlineClients(client *Client) {
	var uniqueUsers = make(map[string]bool)
	for _, user := range h.users {
		if ok := uniqueUsers[user.ID]; !ok {
			message := &entity.Message{
				Action:   config.UserJoinedAction,
				SenderID: user.ID,
			}
			uniqueUsers[user.ID] = true
			client.send <- message.Encode()
		}
	}
}

// Listen to pub/sub general channels
func (h *Hub) listenPubSubChannel() {
	pubsub := h.pubsubRepo.Subscribe(context.Background(), config.PubSubGeneralChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var message entity.Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Println(err)
			continue
		}

		switch message.Action {
		case config.UserJoinedAction:
			h.handleUserJoined(message)
		case config.UserLeftAction:
			h.handleUserLeft(message)
		case config.JoinRoomPrivateAction:
			h.handleUserJoinPrivate(message)
		}
	}
}

func (h *Hub) handleUserJoined(message entity.Message) {
	user, err := h.userRepo.Get(context.Background(), message.SenderID)
	if err != nil {
		log.Println(err)
		return
	}
	h.users = append(h.users, user)
	h.broadcastToClients(message.Encode())
}

func (h *Hub) handleUserLeft(message entity.Message) {
	for i, user := range h.users {
		if user.ID == message.SenderID {
			h.users = append(h.users[:i], h.users[i+1:]...)
			break
		}
	}
	h.broadcastToClients(message.Encode())
}

func (h *Hub) handleUserJoinPrivate(message entity.Message) {
	targetClients := h.findClientsByID(message.Content)
	for _, targetClient := range targetClients {
		room := h.findRoomByID(message.TargetID)
		client := h.findClientByID(message.SenderID)
		targetClient.joinRoom(room.Name, client)
	}
}

func (h *Hub) findClientsByID(ID string) []*Client {
	var foundClients []*Client
	for client := range h.clients {
		if client.ID == ID {
			foundClients = append(foundClients, client)
		}
	}

	return foundClients
}
