package http

import (
	"github.com/ustackq/indagate/pkg/service"
	"go.uber.org/zap"
)

type BucketBackend struct {
	Logger *zap.Logger

	BucketService              service.BucketService
	UserResourceMappingService service.UserResourceMapping
	UserService                service.UserService
	OrganizationService        service.OrganizationService
}
