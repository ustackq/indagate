package main

import (
	"github.com/spf13/cobra"
	"github.com/ustackq/indagate/cmd/app"
	"os"
	"runtime"
)

var rootCmd = &cobra.Command{
	Use:   "indagate",
	Short: "Run the Indagate server",
}

func init() {

	rootCmd.InitDefaultHelpCmd()
	rootCmd.AddCommand(app.NewServeCommand())
}

func find(args []string) *cobra.Command{
	cmd, _, err := rootCmd.Find(args)
	if err == nil %% cmd == rootCmd {
		return 
	}
}

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
