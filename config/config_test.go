package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const testConfigStr = `
db: "postgres://postgres:postgres@localhost:5432/example?sslmode=disable"
listen_addr: "localhost:8000"
sources:
  - name: site1
    url: site1.com
    country: USA
    timeout: 10s
    update: 3600s
    tags:
        country: ["USA"]
        topic: ["sports", "politics"]
        language: ["en"]
        extra: ["test"]

  - name: site2
    url: site2.com
    country: UK
    timeout: 20s
    update: 1800s
    tags:
        country: ["UK"]
        topic: ["tech", "IT", "sports"]
        language: ["en", "ge"]
`

func TestParseConfig(t *testing.T) {
	config, err := parse([]byte(testConfigStr))
	assert.Nil(t, err)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/example?sslmode=disable", config.DbConnString)
	assert.Equal(t, "localhost:8000", config.ListenAddr)
	assert.Len(t, config.Sources, 2)
	assert.Equal(t, "site1", config.Sources[0].Name)
	assert.Equal(t, "site2", config.Sources[1].Name)
	assert.Equal(t, time.Minute*60, config.Sources[0].UpdatePeriod)
}

func TestGetAllTags(t *testing.T) {
	config, err := parse([]byte(testConfigStr))
	assert.Nil(t, err)
	tags := config.GetAllTags()
	assert.Len(t, tags, 4)
	keys := maps.Keys(tags)
	slices.Sort(keys)
	assert.Equal(t, []string{"country", "extra", "language", "topic"}, keys)
	assert.Equal(t, []string{"test"}, tags["extra"])
	topics := tags["topic"]
	slices.Sort(topics)
	assert.Equal(t, []string{"IT", "politics", "sports", "tech"}, topics)
}

func TestMakeTagStringIntoMap(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		result, err := makeTagStringIntoMap([]string{"key1:val1", "key2:val2", "key1:val3"})
		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, []string{"val2"}, result["key2"])
		assert.Equal(t, []string{"val1", "val3"}, result["key1"])
	})

	t.Run("with error", func(t *testing.T) {
		_, err := makeTagStringIntoMap([]string{"key1:val1", "abc", "key1:val3:"})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "has to be in 'tagCategory:tagValue' format")
	})
}

func TestSliceIsSubset(t *testing.T) {
	assert.True(t, sliceIsSubset([]string{"b", "c"}, []string{"a", "b", "c"}))
	assert.False(t, sliceIsSubset([]string{"a", "b", "c"}, []string{"b", "a", "d", "e"}))
}

func TestTagsAreSubset(t *testing.T) {
	tags1 := map[string][]string{"k1": {"v1", "v2", "v3"}, "k2": {"v4", "v5"}}
	tags2 := map[string][]string{"k1": {"v2"}}
	tags3 := map[string][]string{"k1": {"v1"}, "k2": {"v5"}}
	tags4 := map[string][]string{"k3": {"v6"}}

	assert.True(t, tagsAreSubset(tags2, tags1))
	assert.True(t, tagsAreSubset(tags3, tags1))
	assert.False(t, tagsAreSubset(tags4, tags1))
	assert.False(t, tagsAreSubset(tags2, tags3))

	t1 := map[string][]string{"language": {"en"}, "topic": {"sports", "politics"}}
	t2 := map[string][]string{"country": {"UK"}, "language": {"en", "ge"}, "topic": {"tech", "IT", "sports"}}
	assert.False(t, tagsAreSubset(t1, t2))
}

func TestGetResourcesWithTags(t *testing.T) {
	config, err := parse([]byte(testConfigStr))
	assert.Nil(t, err)

	resources, err := config.GetResourcesWithTags([]string{"extra:test"})
	assert.Nil(t, err)
	assert.Equal(t, []string{"site1"}, resources)

	resources, err = config.GetResourcesWithTags([]string{"language:en", "topic:sports"})
	assert.Nil(t, err)
	slices.Sort(resources)
	assert.Equal(t, []string{"site1", "site2"}, resources)

	resources, err = config.GetResourcesWithTags([]string{"language:en", "topic:sports", "topic:politics"})
	fmt.Println("RESOURCES", resources)
	assert.Nil(t, err)
	assert.Equal(t, []string{"site1"}, resources)
}
