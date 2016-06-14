package config

import (
	"github.com/dtan4/paus-frontend/util"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

const (
	ConfigPrefix = "paus"
)

type Config struct {
	BaseDomain         string `envconfig:"base_domain"   required:"true"`
	EtcdEndpoint       string `envconfig:"etcd_endpoint" default:"http://localhost:2379"`
	GitHubClientID     string `envconfig:"github_client_id" required:"true"`
	GitHubClientSecret string `envconfig:"github_client_secret" required:"true"`
	ReleaseMode        bool   `envconfig:"release_mode"  default:"false"`
	SecretKeyBase      string `envconfig:"secret_key_base"`
	SkipKeyUpload      bool   `envconfig:"skip_key_upload" default:"false"`
	URIScheme          string `envconfig:"uri_scheme"    default:"http"`
}

func LoadConfig() (*Config, error) {
	var config Config

	err := envconfig.Process(ConfigPrefix, &config)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to load config from envs.")
	}

	if config.SecretKeyBase == "" {
		s, err := util.GenerateRandomString()

		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate secret key base.")
		}

		config.SecretKeyBase = s
	}

	return &config, nil
}
