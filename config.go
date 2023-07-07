package main

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
	DbConnString string         `yaml:"db"`
	ListenAddr   string         `yaml:"listen_addr"`
	Sources      []SourceConfig `yaml:"sources,flow"`
}

func parseConfig(configBytes []byte) (Config, error) {
	var config Config

	err := yaml.Unmarshal(configBytes, &config)
	return config, err
}

func getConfig(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	return parseConfig(data)
}
