package entity

type User struct {
	// The actual websocket connection.
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
