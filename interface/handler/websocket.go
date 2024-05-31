package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tusmasoma/simple-chat/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebsocketHandler struct {
	hub    repository.HubWebSocketRepository
	client repository.ClientWebSocketRepository
}

func NewWebsocketHandler(hub repository.HubWebSocketRepository, client repository.ClientWebSocketRepository) *WebsocketHandler {
	return &WebsocketHandler{
		hub:    hub,
		client: client,
	}
}

func (h *WebsocketHandler) WebSocketConnection(w http.ResponseWriter, r *http.Request) {

	name, ok := r.URL.Query()["name"]
	if !ok || len(name) < 1 {
		http.Error(w, "Url Param 'name' is missing", http.StatusBadRequest)
		return
	}

	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go h.client.WritePump()
	go h.client.ReadPump()

	//hub.register <- client
}
