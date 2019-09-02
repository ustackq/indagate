package service

import (
	"context"
)

// Organization define org in indagate
// More info: TODO
type Organization struct {
	ID          ID     `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OrganizationFilter represents a set of filter for search
type OrganizationFilter struct {
	Name *string
	ID   *ID
}

// OrganizationService define the service for managing organization data.
type OrganizationService interface {
	FindOrganizationByID(ctx context.Context, id ID) (*Organization, error)
	// FindOrganization return the first org that matches filter
	FindOrganization(ctx context.Context, filter OrganizationFilter) (*Organization, error)
	FindOrganizations(ctx context.Context, filter OrganizationFilter, opt ...FindOptions) ([]*Organization, int, error)
	// CreateOrganization creates a new org
	CreateOrganization(ctx context.Context, b *Organization) error
	// Updates orgnization and returns the new organization state after update.
	UpdateOrganization(ctx context.Context, id ID, update OrganizationUpdate) (*Organization, error)
	DeleteOrganization(ctx context.Context, id ID) error
}

type OrganizationUpdate struct {
	Name        *string
	Description *string `json:"description,omitempty"`
}
