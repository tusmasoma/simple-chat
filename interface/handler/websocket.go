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

type WebsocketHandler struct{}

func NewWebsocketHandler() *WebsocketHandler {
	return &WebsocketHandler{}
}

func (h *WebsocketHandler) WebSocketConnection(hub repository.HubWebSocketRepository, w http.ResponseWriter, r *http.Request, pubsubRepo repository.PubSubRepository) {

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

	//client := ws.NewClientWebSocketRepository(conn, hub)

	//go client.WritePump()
	//go client.ReadPump()

	//hub.register <- client
}
