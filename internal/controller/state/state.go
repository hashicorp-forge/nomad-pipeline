package state

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/state/dev"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/state/nvar"
)

const (
	BackendDev       = "dev"
	BackendNomadVars = "nomad-vars"
)

type Config struct {
	Backend string
	NVars   *NomadVarsConfig
}

type NomadVarsConfig struct {
	CacheEnabled bool
}

func (c *Config) Merge(other *Config) *Config {
	if c == nil {
		return other
	}

	result := *c

	if other.Backend != "" {
		result.Backend = other.Backend
	}

	if other.NVars != nil {
		if result.NVars == nil {
			result.NVars = &NomadVarsConfig{}
		}
		if other.NVars.CacheEnabled {
			result.NVars.CacheEnabled = other.NVars.CacheEnabled
		}
	}

	return &result
}

func (c *Config) Validate() error {
	switch c.Backend {
	case BackendDev, BackendNomadVars:
		return nil
	default:
		return fmt.Errorf("unsupported state backend: %s", c.Backend)
	}
}

func DefaultConfig() *Config {
	return &Config{
		Backend: BackendDev,
		NVars: &NomadVarsConfig{
			CacheEnabled: true,
		},
	}
}

func NewBackend(cfg *Config, logger *zap.Logger, client *api.Client) (state.State, error) {
	switch cfg.Backend {
	case BackendDev:
		return dev.New(), nil
	case BackendNomadVars:
		return nvar.New(cfg.NVars.CacheEnabled, logger, client)
	default:
		panic("not implemented")
	}
}
