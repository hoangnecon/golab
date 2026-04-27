package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds server configuration loaded from environment variables.
type Config struct {
	ColabBaseURL string        // COLAB_BASE_URL (default: colab.research.google.com)
	WSPort       int           // COLAB_WS_PORT (default: 0 = random)
	Token        string        // COLAB_TOKEN   (default: "" = random)
	Timeout      time.Duration // COLAB_TIMEOUT  (default: 120s)
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	c := &Config{
		ColabBaseURL: "colab.research.google.com",
		WSPort:       0,
		Token:        "",
		Timeout:      120 * time.Second,
	}

	if v := os.Getenv("COLAB_BASE_URL"); v != "" {
		c.ColabBaseURL = v
	}
	if v := os.Getenv("COLAB_WS_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.WSPort = p
		}
	}
	if v := os.Getenv("COLAB_TOKEN"); v != "" {
		c.Token = v
	}
	if v := os.Getenv("COLAB_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Timeout = d
		}
	}
	return c
}
