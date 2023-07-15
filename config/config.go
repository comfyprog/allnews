package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type SourceConfig struct {
	Name         string              `yaml:"name"`
	FeedUrl      string              `yaml:"url"`
	Timeout      time.Duration       `yaml:"timeout"`
	UpdatePeriod time.Duration       `yaml:"update"`
	Tags         map[string][]string `yaml:"tags"`
}

type Config struct {
	Version      string         `yaml:"-"`
	DbConnString string         `yaml:"db"`
	ListenAddr   string         `yaml:"listen_addr"`
	Sources      []SourceConfig `yaml:"sources,flow"`
}

func (c Config) GetAllTags() map[string][]string {
	rawTags := make(map[string]map[string]struct{})

	for _, s := range c.Sources {
		for tagCat := range s.Tags {
			if _, ok := rawTags[tagCat]; !ok {
				rawTags[tagCat] = make(map[string]struct{})
			}

			for _, tagVal := range s.Tags[tagCat] {
				rawTags[tagCat][tagVal] = struct{}{}
			}
		}
	}

	tags := make(map[string][]string)
	for tagCat := range rawTags {
		if _, ok := tags[tagCat]; !ok {
			tags[tagCat] = make([]string, 0, len(rawTags[tagCat]))
		}

		for tagVal := range rawTags[tagCat] {
			tags[tagCat] = append(tags[tagCat], tagVal)
		}
	}

	return tags
}

func makeTagStringIntoMap(ts []string) (map[string][]string, error) {
	result := make(map[string][]string)
	for _, s := range ts {
		splitted := strings.Split(s, ":")
		if len(splitted) != 2 {
			return nil, fmt.Errorf("%q has to be in 'tagCategory:tagValue' format", s)
		}
		tagCat, tagVal := splitted[0], splitted[1]
		if _, ok := result[tagCat]; !ok {
			result[tagCat] = []string{tagVal}
		} else {
			result[tagCat] = append(result[tagCat], tagVal)
		}
	}

	return result, nil
}

func sliceIsSubset(subset []string, superset []string) bool {
	for _, v := range subset {
		if !slices.Contains(superset, v) {
			return false
		}
	}

	return true
}

func tagsAreSubset(subset map[string][]string, superset map[string][]string) bool {
	for key := range subset {
		vals2, ok := superset[key]
		if !ok {
			return false
		}

		vals1 := subset[key]

		if !sliceIsSubset(vals1, vals2) {
			return false
		}
	}

	return true
}

func (c Config) GetResourcesWithTags(tagStrings []string) ([]string, error) {
	resources := make([]string, 0)

	tags, err := makeTagStringIntoMap(tagStrings)
	if err != nil {
		return resources, err
	}

	for _, r := range c.Sources {
		if tagsAreSubset(tags, r.Tags) {
			resources = append(resources, r.Name)
		}
	}

	return resources, nil
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
