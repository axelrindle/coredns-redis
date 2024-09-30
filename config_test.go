package redis

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	golog "log"

	"github.com/coredns/caddy"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	block := `
    redis {
        address localhost:6379
    }
    `

	c := caddy.NewTestController("dns", block)
	config, err := parseConfig(c)

	assert.Nil(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "localhost:6379", config.RedisAddress)
	assert.Equal(t, "", config.RedisPassword)
	assert.Equal(t, uint64(0), config.RedisDatabase)

	assert.Equal(t, uint64(30), config.ConnectTimeout)
	assert.Equal(t, uint64(5), config.ReadTimeout)

	assert.Equal(t, "", config.KeyPrefix)
	assert.Equal(t, "", config.KeySuffix)

	assert.Equal(t, uint64(300), config.TTL)
}

func TestConfigInvalid(t *testing.T) {
	block := `
    redis {
        address localhost:6379
        database
    }
    `

	c := caddy.NewTestController("dns", block)
	config, err := parseConfig(c)

	assert.Nil(t, config)
	assert.NotNil(t, err)
}

func TestConfigWarning(t *testing.T) {
	buf := new(bytes.Buffer)
	scanner := bufio.NewScanner(buf)

	golog.SetOutput(buf)

	block := `
    redis {
        address localhost:6379
        database not_a_number
    }
    `

	c := caddy.NewTestController("dns", block)
	config, err := parseConfig(c)

	assert.Nil(t, err)
	assert.NotNil(t, config)

	scanner.Scan()
	line := scanner.Text()
	assert.True(t, strings.Contains(line, "Error parsing config value database"))
	assert.False(t, scanner.Scan())
}
