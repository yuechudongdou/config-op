package cli

import (
	"github.com/configUpdate/cli/config"
	"github.com/configUpdate/cli/migration"
	"github.com/configUpdate/cli/tags"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "configOp",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var gitlabUrl string
var accessToken string

func init() {
	RootCmd.PersistentFlags().StringP(
		"address",
		"a",
		"http://172.16.211.1",
		"gitlab 地址",
	)

	RootCmd.PersistentFlags().StringP(
		"token",
		"t",
		"5dszj_WeFVusrCtVrLZA",
		"gitlab 私有token",
	)



	RootCmd.AddCommand(config.ConfigCmd)
	RootCmd.AddCommand(tags.TagsCli)
	RootCmd.AddCommand(migration.MigrateCmd)
}
