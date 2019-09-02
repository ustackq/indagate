package app

import (
	"context"
	"fmt"
	"os"

	"sync"

	"github.com/coreos/go-systemd/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ustackq/indagate/cmd/app/options"
	"github.com/ustackq/indagate/pkg/signal"
	"github.com/ustackq/indagate/pkg/telemetry"
	"time"
)

func NewServeCommand() *cobra.Command {
	// TODO: add flags
	cfg := viper.GetString("config")
	// parse config
	ing := options.NewIndagateOptions(cfg)
	ing.Parse(cfg)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the indagate server (default)",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			ctx = signals.WithStandardSignals(ctx)
			// handl serve
			if err := ing.Run(ctx); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			// handle telemetry
			var wg sync.WaitGroup
			if ing.TelemetryEnabled() {
				reporter := telemetry.NewTelemetry(ing.Registry())
				reporter.Duration = 12 * time.Hour
				reporter.Logger = ing.Logger
				wg.Add(1)
				go func() {
					defer wg.Done()
					reporter.Report(ctx)
				}()
			}
			<-ctx.Done()

			// handle signal
			go daemon.SdNotify(false, daemon.SdNotifyReady)
			// attempt clean shutdown
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			ing.Shutdown(ctx)
			wg.Wait()

		},
	}
	// init context
	return cmd
}
