package store

import (
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/generator"
	"go.uber.org/zap"
	"time"
)

type Service struct {
	store          Store
	Logger         *zap.Logger
	Hash           *service.BCrypt
	IDGenerator    generator.IDGenerator
	TokenGenerator generator.TokenGenerator
	time           func() time.Time
}

func NewService(s Store) *Service {
	return &Service{
		time:           time.Now,
		Logger:         zap.NewNop(),
		IDGenerator:    generator.NewIDGenerator(),
		TokenGenerator: generator.NewTokenGenerator(64),
		Hash:           &service.BCrypt{},
		store:          s,
	}
}
