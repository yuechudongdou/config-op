package migration

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
)


var downCli = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {
		updateDatabase(cmd, args)
	},
	Args: cobra.MaximumNArgs(1),
}

func downDatabase(cmd *cobra.Command, args []string) {

	limit := -1
	var err error
	if len(args) > 0 {

		limit, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(args[0], " is not a int")
			os.Exit(1)
		}
	}
	if source == "" && pathDir != "" {
		source = fmt.Sprintf("file://%v", pathDir)
	}

	// initialize migrate
	// don't catch migraterErr here and let each command decide
	// how it wants to handle the error
	migrater, migraterErr := migrate.New(source, databaseUrl)
	defer func() {
		if migraterErr == nil {
			if _, err := migrater.Close(); err != nil {
				log.Println(err)
			}
		}
	}()
	if migraterErr != nil {
		fmt.Println(migraterErr)
		os.Exit(1)
	}

	if limit >= 0 {
		if err := migrater.Steps(-limit); err != nil {
			if err != migrate.ErrNoChange {
				os.Exit(1)
			}
			log.Println(err)
		}
	} else {
		if err := migrater.Down(); err != nil {
			if err != migrate.ErrNoChange {
				os.Exit(1)
			}
			log.Println(err)
		}
	}
}