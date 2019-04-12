package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/ustackq/indagate/cmd/app"
	"github.com/ustackq/indagate/pkg/server"
	"go.uber.org/zap"
)

const (
	logLevelKey = "indagate"
)

func main() {
	// consider runtime library usage
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}
	// consider the reason why using it?
	rand.Seed(time.Now().UnixNano())

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	defer logger.Sync()
	command := app.NewIndagateCommand(server.SetupSingalHandler(), logger)

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
