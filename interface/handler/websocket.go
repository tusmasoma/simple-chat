package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tusmasoma/simple-chat/repository"
	"github.com/tusmasoma/simple-chat/usecase"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebsocketHandler struct {
	hub    repository.HubWebSocketRepository
	client repository.ClientWebSocketRepository
	auc    usecase.AuthUseCase
}

func NewWebsocketHandler(hub repository.HubWebSocketRepository, client repository.ClientWebSocketRepository) *WebsocketHandler {
	return &WebsocketHandler{
		hub:    hub,
		client: client,
	}
}

func (h *WebsocketHandler) WebSocketConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := h.auc.GetUserFromContext(ctx)

	if _, err = upgrader.Upgrade(w, r, nil); err != nil {
		log.Println(err)
		return
	}

	// TODO: clientの初期化

	go h.client.WritePump()
	go h.client.ReadPump()

	//hub.register <- client
}
