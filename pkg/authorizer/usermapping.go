package authorizer

import (
	"context"
	"github.com/ustackq/indagate/pkg/service"
)

type OrganizationService interface {
	FindResourceOrganizationID(ctx context.Context, rt service.ResourceType, id service.ID) (service.ID, error)
}

type UserMappingService struct {
	urmService service.UserResourceMappingService
	orgService OrganizationService
}

func NewUserMappingService(orgSVC OrganizationService, urm service.UserResourceMappingService) *UserMappingService {
	return &UserMappingService{
		urmService: urm,
		orgService: orgSVC,
	}
}

func newURMPermission(action service.Action, rt service.ResourceType, orgID, id service.ID) (*service.Permission, error) {
	return service.NewPermissionAtID(id, action, rt, orgID)
}

func authorizaURMByAction(ctx context.Context, rt service.ResourceType, orgID, id service.ID, action service.Action) error {
	p, err := newURMPermission(action, rt, orgID, id)
	if err != nil {
		return err
	}

	if err := isAllowed(ctx, *p); err != nil {
		return err
	}

	return nil
}

// TODO: change int  to int64
func (s *UserMappingService) FindUserResourceMappings(ctx context.Context, filter service.UserResourceMappingFilter, opts ...service.FindOptions) ([]*service.UserResourceMapping, int, error) {
	urms, _, err := s.urmService.FindUserResourceMappings(ctx, filter, opts...)
	if err != nil {
		return nil, 0, err
	}

	mappings := urms[:0]
	for _, urm := range urms {
		orgID, err := s.orgService.FindResourceOrganizationID(ctx, urm.ResourceType, urm.ResourceID)
		if err != nil {
			return nil, 0, err
		}

		if err := authorizaURMByAction(ctx, urm.ResourceType, orgID, urm.ResourceID, service.ReadAction); err != nil {
			continue
		}

		mappings = append(mappings, urm)
	}

	return mappings, len(mappings), nil
}

func (s *UserMappingService) CreateUserResourceMapping(ctx context.Context, m *service.UserResourceMapping) error {
	orgID, err := s.orgService.FindResourceOrganizationID(ctx, m.ResourceType, m.ResourceID)
	if err != nil {
		return err
	}

	if err := authorizaURMByAction(ctx, m.ResourceType, orgID, m.ResourceID, service.WriteAction); err != nil {
		return err
	}

	return s.urmService.CreateUserResourceMapping(ctx, m)
}

func (s *UserMappingService) DeleteUserResourceMapping(ctx context.Context, resourceID service.ID, userID service.ID) error {
	urmf := service.UserResourceMappingFilter{
		ResourceID: resourceID,
		UserID:     userID,
	}

	urms, _, err := s.urmService.FindUserResourceMappings(ctx, urmf)
	if err != nil {
		return err
	}

	for _, urm := range urms {
		orgID, err := s.orgService.FindResourceOrganizationID(ctx, urm.ResourceType, urm.ResourceID)
		if err != nil {
			return err
		}

		if err := authorizaURMByAction(ctx, urm.ResourceType, orgID, urm.ResourceID, service.WriteAction); err != nil {
			return err
		}

		if err := s.urmService.DeleteUserResourceMapping(ctx, urm.ResourceID, urm.UserID); err != nil {
			return err
		}
	}
	return nil
}
