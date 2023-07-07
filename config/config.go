package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type SourceConfig struct {
	Name         string        `yaml:"name"`
	FeedUrl      string        `yaml:"url"`
	Timeout      time.Duration `yaml:"timeout"`
	UpdatePeriod time.Duration `yaml:"update"`
	Country      string        `yaml:"country"`
}

type Config struct {
	Version      string         `yaml:"-"`
	DbConnString string         `yaml:"db"`
	ListenAddr   string         `yaml:"listen_addr"`
	Sources      []SourceConfig `yaml:"sources,flow"`
}

func parse(configBytes []byte) (Config, error) {
	var config Config

	err := yaml.Unmarshal(configBytes, &config)
	return config, err
}

func Get(filename string, appVersion string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	config, err := parse(data)
	if err != nil {
		return config, err
	}

	config.Version = appVersion
	return config, nil
}
