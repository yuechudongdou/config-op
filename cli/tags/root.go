package tags

import (
	"github.com/spf13/cobra"
)

var TagsCli = &cobra.Command{
	Use: "tags",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	TagsCli.AddCommand(createCli)
}