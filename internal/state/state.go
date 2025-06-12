package state

import (
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/state/dev"
)

type Config struct {
	Dev *DevConfig
}

type DevConfig struct {
	Enabled bool
}

func DefaultConfig() *Config {
	return &Config{
		Dev: &DevConfig{
			Enabled: true,
		},
	}
}

func NewBackend(cfg *Config) state.State {
	if cfg.Dev != nil && cfg.Dev.Enabled {
		return dev.New()
	}
	panic("not implemented")
}
