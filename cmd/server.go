package main

import (
	"context"
	"fmt"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

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

	s.notifyClientJoined(client)
	s.listOnlineClients(client)
	s.clients[client] = true

	s.users = append(s.users, user)
}

func (s *WsServer) unregisterClient(client *Client) {
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		s.notifyClientLeft(client)

		user := client.ToUser()
		for i, u := range s.users {
			if u.ID == user.ID {
				s.users = append(s.users[:i], s.users[i+1:]...)
				break
			}
		}
		s.userRepo.RemoveUser(context.Background(), *user)
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
