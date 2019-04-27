package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ustackq/indagate/cmd/app/options"
	"github.com/ustackq/indagate/pkg/telemetry"
	"github.com/ustackq/indagate/pkg/utils/flag"
	"github.com/ustackq/indagate/pkg/version"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Indagate server",
	RunE:  serverCmdF,
}

func init() {
	RootCmd.AddCommand(serverCmd)
	RootCmd.RunE = serverCmdF
}

func serverCmdF(cmd *cobra.Command, args []string) error {
	// flag --version
	version.PrintAndExitIfRequested()
	// print flag and value
	flag.PrintFlags(cmd.Flags())

	cfg := viper.GetString("config")

	disable, err := cmd.Flags().GetBool("configWacher")
	if err != nil {
		return err
	}
	// parse config
	ing := options.NewIndagateOptions(cfg)
	ing.Parse(cfg)
	// construct indagate
	interChan := make(chan os.Signal, 1)

	return runServer(ing, disable, interChan)
}

func runServer(ing *options.Indagate, disable bool, interChan chan os.Signal) error {
	// Learn context usage
	// https://sourcegraph.com/github.com/influxdata/influxdb/-/commit/e8045ae187702eccc6ef2529e0793f3f0ffc1092
	ctx := context.Background()
	defer ing.Shutdown(ctx)
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
	signal.Notify(interChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interChan
	return nil
}
