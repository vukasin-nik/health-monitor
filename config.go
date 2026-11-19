package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
)

// Target is a single HTTP endpoint that is checked periodically.
type Target struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Config is loaded from a JSON file (see targets.json).
type Config struct {
	IntervalSeconds int      `json:"interval_seconds"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	Targets         []Target `json:"targets"`
}

// LoadConfig reads and validates the configuration file and applies defaults.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 10
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 5
	}
	if len(cfg.Targets) == 0 {
		return nil, errors.New("config has no targets")
	}

	for _, t := range cfg.Targets {
		if t.Name == "" {
			return nil, errors.New("target is missing a name")
		}
		u, err := url.Parse(t.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return nil, fmt.Errorf("target %q has an invalid url %q", t.Name, t.URL)
		}
	}

	return &cfg, nil
}
