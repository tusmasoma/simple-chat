package repository

type ClientWebSocketRepository interface {
	ReadPump()
	WritePump()
}
