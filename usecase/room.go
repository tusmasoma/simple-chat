package usecase

type RoomUseCase interface{}

type roomUseCase struct{}

func NewRoomUseCase() RoomUseCase {
	return &roomUseCase{}
}

func (h *Hub) findRoomByID(id string) *Room {
	for room := range h.rooms {
		if room.ID.String() == id {
			return room
		}
	}
	return nil
}
