package app

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ustackq/indagate/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Indagate server",
	RunE:  serverCmdF,
}

func serverCmdF(cmd *cobra.Command, args []string) error {
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
	return nil
}
