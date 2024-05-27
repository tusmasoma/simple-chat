package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

const PubSubGeneralChannel = "general"

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
	users      []*entity.User
	roomRepo   repository.RoomRepository
	userRepo   repository.UserRepository
	pubsubRepo repository.PubSubRepository
}

// NewWebsocketServer creates a new WsServer type
func NewHub(ctx context.Context, roomRepo repository.RoomRepository, userRepo repository.UserRepository, pubsubRepo repository.PubSubRepository) *Hub {
	wsServer := &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
		roomRepo:   roomRepo,
		userRepo:   userRepo,
		pubsubRepo: pubsubRepo,
	}

	wsServer.users = userRepo.GetAllUsers(ctx)

	return wsServer
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
	user := client.ToUser()
	h.userRepo.AddUser(context.Background(), *user)

	h.publishClientJoined(client)

	h.listOnlineClients(client)
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)

		h.userRepo.RemoveUser(context.Background(), *client.ToUser())

		h.publishClientLeft(client)
	}
}

func (h *Hub) broadcastToClients(message []byte) {
	for client := range h.clients {
		client.send <- message
	}
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
	roomEntity := h.roomRepo.FindRoomByName(context.Background(), name)
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
		if room.ID.String() == id {
			return room
		}
	}
	return nil
}

func (h *Hub) findClientByID(id string) *Client {
	for client := range h.clients {
		if client.ID.String() == id {
			return client
		}
	}
	return nil
}

func (h *Hub) createRoom(name string, private bool) *Room {
	room := NewRoom(name, private, h.pubsubRepo)

	h.roomRepo.AddRoom(context.Background(), entity.Room{
		ID:      room.ID,
		Name:    room.Name,
		Private: room.Private,
	})

	go room.Run()
	h.rooms[room] = true
	return room
}

func (h *Hub) notifyClientJoined(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action:  UserJoinedAction,
		Message: fmt.Sprintf(welcomeMessage, client.Name),
		Sender:  user,
	}
	h.broadcastToClients(message.encode())
}

func (h *Hub) notifyClientLeft(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action:  UserLeftAction,
		Message: fmt.Sprintf(goodbyeMessage, client.Name),
		Sender:  user,
	}
	h.broadcastToClients(message.encode())
}

func (h *Hub) listOnlineClients(client *Client) {
	for _, user := range h.users {
		message := &Message{
			Action: UserJoinedAction,
			Sender: user,
		}
		client.send <- message.encode()
	}
}

// PublishClientJoined publishes a message to the general channel when a client joins the server
func (h *Hub) publishClientJoined(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action: UserJoinedAction,
		Sender: user,
	}
	if err := h.pubsubRepo.Publish(context.Background(), PubSubGeneralChannel, message.encode()); err != nil {
		log.Println(err)
	}
}

// PublishClientLeft publishes a message to the general channel when a client leaves the server
func (h *Hub) publishClientLeft(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action: UserLeftAction,
		Sender: user,
	}
	if err := h.pubsubRepo.Publish(context.Background(), PubSubGeneralChannel, message.encode()); err != nil {
		log.Println(err)
	}
}

// Listen to pub/sub general channels
func (h *Hub) listenPubSubChannel() {
	pubsub := h.pubsubRepo.Subscribe(context.Background(), PubSubGeneralChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var message Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Println(err)
			continue
		}

		switch message.Action {
		case UserJoinedAction:
			h.handleUserJoined(message)
		case UserLeftAction:
			h.handleUserLeft(message)
		case JoinRoomPrivateAction:
			h.handleUserJoinPrivate(message)
		}
	}
}

func (h *Hub) handleUserJoined(message Message) {
	h.users = append(h.users, message.Sender)
	h.broadcastToClients(message.encode())
}

func (h *Hub) handleUserLeft(message Message) {
	for i, user := range h.users {
		if user.ID == message.Sender.ID {
			h.users = append(h.users[:i], h.users[i+1:]...)
			break
		}
	}
	h.broadcastToClients(message.encode())
}

func (h *Hub) handleUserJoinPrivate(message Message) {
	target := h.findClientByID(message.Message)
	if target != nil {
		target.joinRoom(message.Target.Name, message.Sender)
	}
}
