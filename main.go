package main

import (
	"fmt"
	"github.com/configUpdate/cli"
	"os"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
