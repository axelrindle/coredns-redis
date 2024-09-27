package redis

import (
	"encoding/json"
	"strconv"

	"github.com/coredns/caddy"
)

type Config struct {
	RedisAddress  string `json:"redis_address"`
	RedisPassword string `json:"-"`
	RedisDatabase uint64 `json:"redis_database"`

	ConnectTimeout uint64 `json:"timeout_connect"`
	ReadTimeout    uint64 `json:"timeout_read"`

	KeyPrefix string `json:"prefix_key"`
	KeySuffix string `json:"suffix_key"`

	TTL uint64 `json:"ttl"`
}

func parseConfig(c *caddy.Controller) (*Config, error) {
	config := &Config{
		RedisDatabase: 0,

		ConnectTimeout: 30,
		ReadTimeout:    5,

		KeyPrefix: "",
		KeySuffix: "",

		TTL: 300,
	}

	var err error

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "address":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.RedisAddress = c.Val()

			case "password":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.RedisPassword = c.Val()

			case "database":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				val, err := strconv.ParseUint(c.Val(), 10, 32)
				if err != nil {
					log.Warningf("Error parsing config value database: %v", err)
				} else {
					config.RedisDatabase = val
				}

			case "prefix":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.KeyPrefix = c.Val()

			case "suffix":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				config.KeySuffix = c.Val()

			case "connect_timeout":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				val, err := strconv.ParseUint(c.Val(), 10, 32)
				if err != nil {
					log.Warningf("Error parsing config value read_timeout: %v", err)
				} else {
					config.ConnectTimeout = val
				}

			case "read_timeout":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				val, err := strconv.ParseUint(c.Val(), 10, 32)
				if err != nil {
					log.Warningf("Error parsing config value read_timeout: %v", err)
				} else {
					config.ReadTimeout = val
				}

			case "ttl":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				val, err := strconv.ParseUint(c.Val(), 10, 32)
				if err != nil {
					log.Warningf("Error parsing config value ttl: %v", err)
				} else {
					config.TTL = val
				}

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
