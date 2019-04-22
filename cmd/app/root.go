package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Command = cobra.Command

func Run(args []string) error {
	return nil
}

var RootCmd = &cobra.Command{
	Use:   "indagate",
	Short: "Open source, self-hosted stackoverflow",
	Long:  "Indagate build communities where everyone who codes can learn and share their knowledge.Documentation available at https://indagate.ustackq.io",
}

func init() {
	// set global flag for each subcommand
	RootCmd.PersistentFlags().StringP("config", "c", "config.json", "Configuration file to parsed.")
	RootCmd.PersistentFlags().Bool("configWacher", false, "When set true, configwatcher will reload when config file changed.")

	// Bind env config
	viper.BindEnv("config")
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
}
