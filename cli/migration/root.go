package migration

import (
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use: "migrate",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var (
	source string
	pathDir        string
	databaseUrl string
	prefetch    int
	lockTimeOut int

)

func init() {
	MigrateCmd.PersistentFlags().StringVarP(
		&source,
		"source",
		"",
		"",
		"Location of the migrations (driver://url)",
	)

	MigrateCmd.PersistentFlags().StringVarP(
		&pathDir,
		"path",
		"",
		"",
		"Shorthand for -source=file://pathDir",
	)

	MigrateCmd.PersistentFlags().StringVarP(
		&databaseUrl,
		"database",
		"",
		"",
		"Run migrations against this databaseUrl (driver://url)",
	)

	MigrateCmd.PersistentFlags().IntVarP(
		&lockTimeOut,
		"lock-timeout",
		"",
		15,
		"Allow N seconds to acquire databaseUrl lock",
	)
	MigrateCmd.PersistentFlags().IntVarP(
		&prefetch,
		"prefetch",
		"",
		10,
		"Number of migrations to load in advance before executing",
	)


	MigrateCmd.AddCommand(upCli)
	MigrateCmd.AddCommand(patchCli)
}

