package store

import (
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/rand"
	"go.uber.org/zap"
	"time"
)

type Service struct {
	store    Store
	Logger   *zap.Logger
	TokenGen service.TokenGenerator
	Hash     Crypt
	time     func() time.Time
}

func NewService(store Store) *Service {
	return &Service{
		store:    store,
		Logger:   zap.NewNop(),
		Hash:     &BCrypt{},
		TokenGen: rand.NewTokenGenerator(64),
	}
}
