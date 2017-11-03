package cmd

import (
	"io/ioutil"

	"io"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	LogLevel string `yaml:"log_level"`

	UAA             uaaConfig `yaml:"uaa"`
	CloudController ccConfig  `yaml:"cloud_controller"`
}

type uaaConfig struct {
	URL        string `yaml:"url"`
	CACertPath string `yaml:"ca_cert_path"`
}

type ccConfig struct {
	URL          string `yaml:"url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func NewConfig(r io.Reader) (Config, error) {
	config := Config{}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(b, &config)
	return config, err
}
