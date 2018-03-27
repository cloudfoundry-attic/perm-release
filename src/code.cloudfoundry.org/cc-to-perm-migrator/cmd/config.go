package cmd

import (
	"errors"
	"io/ioutil"

	"io"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Logger LagerConfig `yaml:",inline"`
	DryRun bool        `yaml:"dry_run"`

	UAA             uaaConfig  `yaml:"uaa"`
	CloudController ccConfig   `yaml:"cloud_controller"`
	Perm            permConfig `yaml:"perm"`
}

type uaaConfig struct {
	URL        string           `yaml:"url"`
	CACertPath FileOrStringFlag `yaml:"ca_cert_path"`
}

type ccConfig struct {
	URL          string   `yaml:"url"`
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	ClientScopes []string `yaml:"client_scopes"`
}

type permConfig struct {
	Hostname string           `yaml:"hostname"`
	Port     int              `yaml:"port"`
	CACert   FileOrStringFlag `yaml:"ca_cert"`
}

func NewConfig(r io.Reader) (*Config, error) {
	config := Config{}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	err = config.validate()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validate() error {
	if len(c.CloudController.ClientScopes) == 0 {
		return errors.New("invalid configuration: must request client scopes")
	}

	return nil
}
