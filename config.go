package redis

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/coredns/caddy"
)

type Config struct {
	redisAddress  string
	redisPassword string

	connectTimeout int

	readTimeout int
	keyPrefix   string
	keySuffix   string

	Ttl            uint32
	Zones          []string
	LastZoneUpdate time.Time
}

func parseConfig(c *caddy.Controller) (*Config, error) {
	config := &Config{
		keyPrefix: "",
		keySuffix: "",
		Ttl:       300,
	}

	var (
		err error
	)

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "address":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.redisAddress = c.Val()
			case "password":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.redisPassword = c.Val()
			case "prefix":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.keyPrefix = c.Val()
			case "suffix":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.keySuffix = c.Val()
			case "connect_timeout":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.connectTimeout, err = strconv.Atoi(c.Val())
				if err != nil {
					config.connectTimeout = 0
				}
			case "read_timeout":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.readTimeout, err = strconv.Atoi(c.Val())
				if err != nil {
					config.readTimeout = 0
				}
			case "ttl":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultTtl
				}
				config.Ttl = uint32(val)
			default:
				if c.Val() != "}" {
					return nil, c.Errf("unknown property '%s'", c.Val())
				}
			}
		}
	}

	configString, err := json.Marshal(config)
	if err == nil {
		log.Debug(string(configString))
	}

	return config, nil
}
