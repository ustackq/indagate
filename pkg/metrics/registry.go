package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Registry wrap prometheus registry
type Registry struct {
	*prometheus.Registry
	logger *zap.Logger
}

func NewRegistry() *Registry {
	return &Registry{
		Registry: prometheus.NewRegistry(),
		logger:   zap.NewNop(),
	}
}

func (reg *Registry) WithLogger(logger *zap.Logger) {
	reg.logger = logger.With(zap.String("service", "registry"))
}
