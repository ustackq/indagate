package service

import (
	"context"
	"errors"
)

var (
	// ErrInvalidUserType notes that the provided UserType is invalid
	ErrInvalidUserType    = errors.New("unknown user type")
	ErrInvalidMappingType = errors.New("unknown mapping type")
)

type UserType string

const (
	// Owner can read and write to a resource
	Owner UserType = "owner"
	// Member can read from a resource.
	Member UserType = "member"
)

func (ut UserType) Valid() (err error) {
	switch ut {
	case Owner:
	case Member:
	default:
		err = ErrInvalidUserType
	}
	return err
}

type MappingType uint8

const (
	UserMappingType = 0
	OrgMappingType  = 1
)

func (mt MappingType) String() string {
	switch mt {
	case UserMappingType:
		return "user"
	case OrgMappingType:
		return "org"
	}

	return "unknown"
}

func (mt MappingType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

func (mt *MappingType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "user":
		*mt = UserMappingType
		return nil
	case "org":
		*mt = OrgMappingType
		return nil
	}
	return ErrInvalidMappingType
}

type UserResourceMapping struct {
	UserID       ID           `json:"userID"`
	UserType     UserType     `json:"userType"`
	MappingType  MappingType  `json:"mappingType"`
	ResourceType ResourceType `json:"resourceType"`
	ResourceID   ID           `json:"resourceID"`
}

type UserResourceMappingFilter struct {
	ResourceID   ID
	ResourceType ResourceType
	UserID       ID
	UserType     UserType
}

type UserResourceMappingService interface {
	// FindUserResourceMappings returns a list of UserResourceMappings that match filter and the total count of matching mappings.
	FindUserResourceMappings(ctx context.Context, filter UserResourceMappingFilter, opt ...FindOptions) ([]*UserResourceMapping, int, error)

	// CreateUserResourceMapping creates a user resource mapping.
	CreateUserResourceMapping(ctx context.Context, m *UserResourceMapping) error

	// DeleteUserResourceMapping deletes a user resource mapping.
	DeleteUserResourceMapping(ctx context.Context, resourceID, userID ID) error
}
