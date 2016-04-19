package main

import (
	"github.com/kelseyhightower/envconfig"
)

const (
	ConfigPrefix = "paus"
)

type Config struct {
	BaseDomain   string `envconfig:"base_domain"   required:"true"`
	EtcdEndpoint string `envconfig:"etcd_endpoint" default:"http://localhost:2379"`
	URIScheme    string `envconfig:"uri_scheme"    default:"http"`
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process("paus", &config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
