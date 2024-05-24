package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tusmasoma/simple-chat/config"
	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

const PubSubGeneralChannel = "general"

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
	users      []*entity.User
	roomRepo   repository.RoomRepository
	userRepo   repository.UserRepository
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer(ctx context.Context, roomRepo repository.RoomRepository, userRepo repository.UserRepository) *WsServer {
	wsServer := &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
		roomRepo:   roomRepo,
		userRepo:   userRepo,
	}

	wsServer.users = userRepo.GetAllUsers(ctx)

	return wsServer
}

// Run starts the server and listens for incoming messages
func (s *WsServer) Run() {
	go s.listenPubSubChannel()

	for {
		select {

		case client := <-s.register:
			s.registerClient(client)

		case client := <-s.unregister:
			s.unregisterClient(client)

		case message := <-s.broadcast:
			s.broadcastToClients(message)
		}

	}
}

func (s *WsServer) registerClient(client *Client) {
	user := client.ToUser()
	s.userRepo.AddUser(context.Background(), *user)

	s.publishClientJoined(client)

	s.listOnlineClients(client)
	s.clients[client] = true
}

func (s *WsServer) unregisterClient(client *Client) {
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)

		s.userRepo.RemoveUser(context.Background(), *client.ToUser())

		s.publishClientLeft(client)
	}
}

func (s *WsServer) broadcastToClients(message []byte) {
	for client := range s.clients {
		client.send <- message
	}
}

func (s *WsServer) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range s.rooms {
		if room.Name == name {
			foundRoom = room
			break
		}
	}

	if foundRoom == nil {
		foundRoom = s.runRoomFromRepository(name)
	}

	return foundRoom
}

func (s *WsServer) runRoomFromRepository(name string) *Room {
	var room *Room
	roomEntity := s.roomRepo.FindRoomByName(context.Background(), name)
	if roomEntity != nil {
		room = NewRoom(roomEntity.Name, roomEntity.Private)
		room.ID = roomEntity.ID

		go room.Run()
		s.rooms[room] = true
	}
	return room
}

func (s *WsServer) findRoomByID(id string) *Room {
	for room := range s.rooms {
		if room.ID.String() == id {
			return room
		}
	}
	return nil
}

func (s *WsServer) findClientByID(id string) *Client {
	for client := range s.clients {
		if client.ID.String() == id {
			return client
		}
	}
	return nil
}

func (s *WsServer) createRoom(name string, private bool) *Room {
	room := NewRoom(name, private)

	s.roomRepo.AddRoom(context.Background(), entity.Room{
		ID:      room.ID,
		Name:    room.Name,
		Private: room.Private,
	})

	go room.Run()
	s.rooms[room] = true
	return room
}

func (s *WsServer) notifyClientJoined(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action:  UserJoinedAction,
		Message: fmt.Sprintf(welcomeMessage, client.Name),
		Sender:  user,
	}
	s.broadcastToClients(message.encode())
}

func (s *WsServer) notifyClientLeft(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action:  UserLeftAction,
		Message: fmt.Sprintf(goodbyeMessage, client.Name),
		Sender:  user,
	}
	s.broadcastToClients(message.encode())
}

func (s *WsServer) listOnlineClients(client *Client) {
	for _, user := range s.users {
		message := &Message{
			Action: UserJoinedAction,
			Sender: user,
		}
		client.send <- message.encode()
	}
}

// PublishClientJoined publishes a message to the general channel when a client joins the server
func (s *WsServer) publishClientJoined(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action: UserJoinedAction,
		Sender: user,
	}
	if err := config.Redis.Publish(context.Background(), PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

// PublishClientLeft publishes a message to the general channel when a client leaves the server
func (s *WsServer) publishClientLeft(client *Client) {
	user := client.ToUser()
	message := &Message{
		Action: UserLeftAction,
		Sender: user,
	}
	if err := config.Redis.Publish(context.Background(), PubSubGeneralChannel, message.encode()).Err(); err != nil {
		log.Println(err)
	}
}

// Listen to pub/sub general channels
func (s *WsServer) listenPubSubChannel() {
	pubsub := config.Redis.Subscribe(context.Background(), PubSubGeneralChannel)
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
			s.handleUserJoined(message)
		case UserLeftAction:
			s.handleUserLeft(message)
		case JoinRoomPrivateAction:
			s.handleUserJoinPrivate(message)
		}
	}
}

func (s *WsServer) handleUserJoined(message Message) {
	s.users = append(s.users, message.Sender)
	s.broadcastToClients(message.encode())
}

func (s *WsServer) handleUserLeft(message Message) {
	for i, user := range s.users {
		if user.ID == message.Sender.ID {
			s.users = append(s.users[:i], s.users[i+1:]...)
			break
		}
	}
	s.broadcastToClients(message.encode())
}

func (s *WsServer) handleUserJoinPrivate(message Message) {
	target := s.findClientByID(message.Message)
	if target != nil {
		target.joinRoom(message.Target.Name, message.Sender)
	}
}
