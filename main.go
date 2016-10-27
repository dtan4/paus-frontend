package main

import (
	"fmt"
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/server"
)

func main() {
	printVersion()

	config, err := config.LoadConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	server.Run(config)
}
