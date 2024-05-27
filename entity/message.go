package entity

type Message struct {
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	ClientID string `json:"client_id"`
	Content  string `json:"content"`
	Action   string `json:"action"`
}
