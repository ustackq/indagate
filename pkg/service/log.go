package service

import (
	"time"
)

// OperationLog is the struct to store crud related ops.
type OperationLog struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TimeGenerator represents a generator for now.
type TimeGenerator interface {
	// Now creates the generated time.
	Now() time.Time
}

// RealTimeGenerator will generate the real time.
type RealTimeGenerator struct{}

// Now returns the current time.
func (g RealTimeGenerator) Now() time.Time {
	return time.Now()
}
