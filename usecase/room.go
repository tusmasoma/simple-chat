package usecase

type RoomUseCase interface{}

type roomUseCase struct{}

func NewRoomUseCase() RoomUseCase {
	return &roomUseCase{}
}
