package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	ConfigPrefix = "paus"
)

type Config struct {
	BaseDomain   string `envconfig:"base_domain"   required:"true"`
	EtcdEndpoint string `envconfig:"etcd_endpoint" default:"http://localhost:2379"`
	ReleaseMode  bool   `envconfig:"release_mode"  default:"false"`
	URIScheme    string `envconfig:"uri_scheme"    default:"http"`
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process("paus", &config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to load config from envs.")
	}

	return &config, nil
}
