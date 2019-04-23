package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type Command = cobra.Command

func Run(args []string) error {
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
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
	RootCmd.PersistentFlags().String("log-level", zapcore.InfoLevel.String(), "Logger levels, now supported:DEBUG, INFO and ERROR.")
	RootCmd.PersistentFlags().Bool("testing", false, "add /debug/info endpoint to clear stores, used for e2e testing.")
	RootCmd.PersistentFlags().Bool("telemetry", false, "disable sending teletmetry data to custom webhook every 4 hours.")

	// Bind env config
	viper.BindEnv("config")
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
}
