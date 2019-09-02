package signals

import (
	"os"
	"os/signal"
	"syscall"

	"context"
)

var onlyOneSignalHandler = make(chan struct{})
var shutdownHandler chan os.Signal

// SetupSingalHandler registryd for SIGTERM and SIGINT.
func SetupSingalHandler() <-chan struct{} {
	close(onlyOneSignalHandler)

	shutdownHandler = make(chan os.Signal, 2)

	stop := make(chan struct{})
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		<-shutdownHandler
		close(stop)
		<-shutdownHandler
		os.Exit(1)
	}()

	return stop
}

func WithStandardSignals(ctx context.Context) context.Context {
	return withSignals(ctx, os.Interrupt, syscall.SIGTERM)
}

func withSignals(ctx context.Context, sigs ...os.Signal) context.Context {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			return
		}
	}()
	return ctx
}
