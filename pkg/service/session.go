package service

import "time"

// Session represents user session
type Session struct {
	ID          ID           `json:"id"`
	Key         string       `json:"key"`
	CreatedAt   time.Time    `json:"createAt"`
	ExpiresAt   time.Time    `json:"expiresAt"`
	UserID      ID           `json:"userID,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

