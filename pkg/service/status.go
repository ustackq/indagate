package service

import (
	"fmt"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

// Status define resource is active or inactive.
type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

// Valid determines if a Status value matches the enum.
func (s Status) Valid() error {
	switch s {
	case Active, Inactive:
		return nil
	default:
		return &errors.Error{
			Code: errors.Invalid,
			Msg:  fmt.Sprintf("invalid status: must be %v or %v", Active, Inactive),
		}
	}
}
