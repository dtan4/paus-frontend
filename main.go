package main

import (
	"os"

	"github.com/dtan4/paus-frontend/config"
	"github.com/dtan4/paus-frontend/server"
	"github.com/dtan4/paus-frontend/store"
	"github.com/pkg/errors"
)

func main() {
	printVersion()

	config, err := config.LoadConfig()

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	etcd, err := store.NewEtcd(config.EtcdEndpoint)

	if err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	if err = etcd.Init(config); err != nil {
		errors.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	server.Run(config, etcd)
}
