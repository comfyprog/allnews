package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	configStr := `
db: "postgres://postgres:postgres@localhost:5432/example?sslmode=disable"
listen_addr: "localhost:8000"
sources:
  - name: site1
    url: site1.com
    country: USA
    timeout: 10s
    update: 3600s

  - name: site2
    url: site2.com
    country: UK
    timeout: 20s
    update: 1800s
`

	config, err := parse([]byte(configStr))
	assert.Nil(t, err)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/example?sslmode=disable", config.DbConnString)
	assert.Equal(t, "localhost:8000", config.ListenAddr)
	assert.Len(t, config.Sources, 2)
	assert.Equal(t, "site1", config.Sources[0].Name)
	assert.Equal(t, "site2", config.Sources[1].Name)
	assert.Equal(t, time.Minute*60, config.Sources[0].UpdatePeriod)
}
