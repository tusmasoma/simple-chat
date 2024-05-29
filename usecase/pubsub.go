package usecase

import (
	"context"
	"log"

	"github.com/tusmasoma/simple-chat/entity"
	"github.com/tusmasoma/simple-chat/repository"
)

const PubSubGeneralChannel = "general"

const (
	SendMessageAction     = "send_message"
	JoinRoomAction        = "join_room"
	LeaveRoomAction       = "leave_room"
	UserJoinedAction      = "user_joined"
	UserLeftAction        = "user_left"
	JoinRoomPrivateAction = "join_room_private"
	RoomJoinedAction      = "room-joined"
)

type PubSubUseCase interface{}

type pubsubUseCase struct {
	clientRepo repository.ClientRepository
	pubsubRepo repository.PubSubRepository
}

func NewPubSubUseCase() PubSubUseCase {
	return &pubsubUseCase{}
}

func (h *pubsubUseCase) publishClientJoined(ctx context.Context, clientID string, clientName string) error {
	if err := h.clientRepo.Create(ctx, entity.Client{
		ID:   clientID,
		Name: clientName,
	}); err != nil {
		log.Println(err)
		return err
	}

	message := &entity.Message{
		Action:   UserJoinedAction,
		ClientID: clientID,
	}
	if err := h.pubsubRepo.Publish(context.Background(), PubSubGeneralChannel, message.Encode()); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// PublishClientLeft publishes a message to the general channel when a client leaves the server
func (h *pubsubUseCase) publishClientLeft(ctx context.Context, clientID string) error {
	if err := h.clientRepo.Delete(ctx, clientID); err != nil {
		log.Println(err)
		return err
	}

	message := &entity.Message{
		Action:   UserLeftAction,
		ClientID: clientID,
	}
	if err := h.pubsubRepo.Publish(context.Background(), PubSubGeneralChannel, message.Encode()); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
