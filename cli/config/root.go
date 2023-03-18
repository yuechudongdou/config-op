package config

import "github.com/spf13/cobra"

var ConfigCmd = &cobra.Command{
	Use: "config",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}
func init() {
	ConfigCmd.AddCommand(updateCli)
}
