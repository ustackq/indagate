package service

// Organization define org in indagate
// More info: TODO
type Organization struct {
	ID          ID     `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
