package service

// User define a user info
type User struct {
	ID   ID     `json:"id,omitempty"`
	Name string `json:"name"`
}
