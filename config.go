package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prananshsingh/rate-limiter-poc/limiter"
)

type Config struct {
	Rules []limiter.Rule `json:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("config has no rules")
	}
	for _, r := range cfg.Rules {
		if r.Name == "" {
			return nil, fmt.Errorf("rule missing name")
		}
		switch r.Scope {
		case "ip", "global", "key":
		default:
			return nil, fmt.Errorf("rule %q: unknown scope %q", r.Name, r.Scope)
		}
		if r.Rate <= 0 {
			return nil, fmt.Errorf("rule %q: rate must be > 0", r.Name)
		}
		if r.Burst < 1 {
			return nil, fmt.Errorf("rule %q: burst must be >= 1", r.Name)
		}
	}

	return &cfg, nil
}
