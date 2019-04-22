package main

import (
	"os"
	"runtime"

	"github.com/ustackq/indagate/cmd/app"
)

func main() {
	// consider runtime library usage
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}
	if err := app.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
