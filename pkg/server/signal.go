package server

import (
	"os"
	"os/signal"
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
