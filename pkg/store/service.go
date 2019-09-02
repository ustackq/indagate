package store

import (
	"context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/generator"
	"go.uber.org/zap"
	"time"
)

type Service struct {
	store          Store
	Logger         *zap.Logger
	Config         ServiceConfig
	Hash           *service.BCrypt
	IDGenerator    service.IDGenerator
	TokenGenerator generator.TokenGenerator
	time           func() time.Time
}

func NewService(s Store, configs ...ServiceConfig) *Service {
	service := &Service{
		time:           time.Now,
		Logger:         zap.NewNop(),
		IDGenerator:    generator.NewIDGenerator(),
		TokenGenerator: generator.NewTokenGenerator(64),
		Hash:           &service.BCrypt{},
		store:          s,
	}
	if len(configs) > 0 {
		service.Config = configs[0]
	} else {
		service.Config.SessionLength = time.Minute * 60
	}
	return service
}

// ServiceConfig allows admin to configure session service.
type ServiceConfig struct {
	SessionLength time.Duration
}

func (s *Service) Init(ctx context.Context) error {

	return s.store.Modify(ctx, func(tx Impl) error {
		if err := s.initializeAuth(ctx, tx); err != nil {
			return err
		}
		// TODO: other service
		return s.initializaUsers(ctx, tx)
	})
}
