package telemetry

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ustackq/indagate/pkg/logger"
	"go.uber.org/zap"
)

type Telemetry struct {
	*Pusher
	Logger   *zap.Logger
	Duration time.Duration
}

func NewTelemetry(g prometheus.Gatherer) *Telemetry {
	return &Telemetry{
		Pusher:   NewPusher(g),
		Logger:   zap.NewNop(),
		Duration: 12 * time.Hour,
	}
}

func (t *Telemetry) Report(ctx context.Context) {
	log := t.Logger.With(
		zap.String("service", "telemetry"),
		logger.DurationLiteral("duration", t.Duration),
	)

	log.Info("Starting")
	if err := t.Push(ctx); err != nil {
		log.Debug("failure reporting indagate metrics", zap.Error(err))
	}
	ticker := time.NewTicker(t.Duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debug("Reporting")
			if err := t.Push(ctx); err != nil {
				log.Debug("failure reporting indagate metrics", zap.Error(err))
			}
		case <-ctx.Done():
			log.Info("Stopping")
			return
		}
	}
}
