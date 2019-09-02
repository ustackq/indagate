package authorizer

import (
	"context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

var _ service.OrganizationService = (*OrgService)(nil)

type OrgService struct {
	s service.OrganizationService
}

func NewOrgService(s service.OrganizationService) *OrgService {
	return &OrgService{
		s: s,
	}
}

func newOrgPermission(action service.Action, id service.ID) (*service.Permission, error) {
	p := &service.Permission{
		Action: action,
		Resource: service.Resource{
			Type: service.OrgsResourceType,
			ID:   &id,
		},
	}
	return p, p.Valid()
}

func authorizeOrgByAction(action service.Action, ctx context.Context, id service.ID) error {
	p, err := newOrgPermission(action, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

func (o *OrgService) FindOrganizationByID(ctx context.Context, id service.ID) (*service.Organization, error) {
	if err := authorizeOrgByAction(service.ReadAction, ctx, id); err != nil {
		return nil, err
	}

	return o.s.FindOrganizationByID(ctx, id)
}

func (o *OrgService) FindOrganization(ctx context.Context, filter service.OrganizationFilter) (*service.Organization, error) {
	org, err := o.s.FindOrganization(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := authorizeOrgByAction(service.ReadAction, ctx, org.ID); err != nil {
		return nil, err
	}

	return org, nil
}

func (o *OrgService) FindOrganizations(ctx context.Context, filter service.OrganizationFilter, opts ...service.FindOptions) ([]*service.Organization, int, error) {
	// TODO: cache results
	os, _, err := o.s.FindOrganizations(ctx, filter, opts...)
	if err != nil {
		return nil, 0, err
	}
	orgs := os[:0]
	for _, o := range os {
		err := authorizeOrgByAction(service.ReadAction, ctx, o.ID)
		if err != nil && errors.ErrorCode(err) != errors.Unauthorized {
			return nil, 0, err
		}

		if errors.ErrorCode(err) == errors.Unauthorized {
			continue
		}

		orgs = append(orgs, o)
	}

	return orgs, len(orgs), nil
}

func (o *OrgService) CreateOrganization(ctx context.Context, org *service.Organization) error {
	p, err := service.NewGlobalPermission(service.WriteAction, service.OrgsResourceType)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return o.s.CreateOrganization(ctx, org)
}

func (o *OrgService) UpdateOrganization(ctx context.Context, id service.ID, update service.OrganizationUpdate) (*service.Organization, error) {
	if err := authorizeOrgByAction(service.WriteAction, ctx, id); err != nil {
		return nil, err
	}

	return o.s.UpdateOrganization(ctx, id, update)
}

func (o *OrgService) DeleteOrganization(ctx context.Context, id service.ID) error {
	if err := authorizeOrgByAction(service.WriteAction, ctx, id); err != nil {
		return err
	}

	return o.s.DeleteOrganization(ctx, id)
}
