package app

import "github.com/spf13/cobra"

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Indagate server",
	RunE:  serverCmdF,
}

func serverCmdF(cmd *cobra.Command, args []string) error {

	return nil
}
