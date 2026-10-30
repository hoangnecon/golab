package config

import (
	"os"
	"strconv"
)

// Config holds Golab daemon configuration loaded from environment variables.
type Config struct {
	UIPort  int
	NoUI    bool
	NoMCP   bool
	Open    bool
	DBPath  string
	CDPPort int
}

// Load reads environment variables with sensible defaults.
func Load() *Config {
	c := &Config{
		UIPort:  9120,
		CDPPort: 9222,
	}

	if v := os.Getenv("GOLAB_UI_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			c.UIPort = p
		}
	}
	if v := os.Getenv("GOLAB_CDP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			c.CDPPort = p
		}
	}
	if v := os.Getenv("GOLAB_DB"); v != "" {
		c.DBPath = v
	}
	return c
}
