package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ustackq/indagate/config"
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

	interChan := make(chan os.Signal, 1)

	configStore, err := config.NewStore(cfg, disable)
	if err != nil {
		return err
	}
	return runServer(configStore, disable, interChan)
}

func runServer(cfg config.Store, disable bool, interChan chan os.Signal) error {

	// handle signal
	go daemon.SdNotify(false, daemon.SdNotifyReady)
	signal.Notify(interChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interChan
	return nil
}
