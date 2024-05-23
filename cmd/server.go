package main

import "fmt"

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
	}
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
	s.notifyClientJoined(client)
	s.listOnlineClients(client)
	s.clients[client] = true
}

func (s *WsServer) unregisterClient(client *Client) {
	if _, ok := s.clients[client]; ok {
		delete(s.clients, client)
		s.notifyClientLeft(client)
	}
}

func (s *WsServer) broadcastToClients(message []byte) {
	for client := range s.clients {
		client.send <- message
	}
}

func (s *WsServer) findRoomByName(name string) *Room {
	for room := range s.rooms {
		if room.Name == name {
			return room
		}
	}
	return nil
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
	go room.Run()
	s.rooms[room] = true
	return room
}

func (s *WsServer) notifyClientJoined(client *Client) {
	message := &Message{
		Action:  UserJoinedAction,
		Message: fmt.Sprintf(welcomeMessage, client.Name),
		Sender:  client,
	}
	s.broadcastToClients(message.encode())
}

func (s *WsServer) notifyClientLeft(client *Client) {
	message := &Message{
		Action:  UserLeftAction,
		Message: fmt.Sprintf(goodbyeMessage, client.Name),
		Sender:  client,
	}
	s.broadcastToClients(message.encode())
}

func (s *WsServer) listOnlineClients(client *Client) {
	for c := range s.clients {
		message := &Message{
			Action:  UserJoinedAction,
			Message: c.Name,
			Sender:  client,
		}
		client.send <- message.encode()
	}
}
