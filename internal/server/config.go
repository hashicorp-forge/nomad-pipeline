package server

import (
	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp-forge/nomad-pipeline/internal/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/state"
)

type Config struct {
	Data  *DataConfig    `hcl:"data,optional"`
	Log   *logger.Config `hcl:"log,optional"`
	HTTP  *HTTPConfig    `hcl:"http,optional"`
	State *state.Config  `hcl:"state,optional"`
}

type DataConfig struct {
	Path string `hcl:"path,optional"`
}

type HTTPConfig struct {
	Binds          []*BindConfig `hcl:"bind,optional"`
	AccessLogLevel string        `hcl:"access_log_level,optional"`
}

type BindConfig struct {
	Addr string `hcl:"addr,optional"`
}

func DefaultConfig() *Config {
	return &Config{
		Data: &DataConfig{
			Path: "/tmp/nomad-pipeline/data",
		},
		Log: logger.DefaultConfig(),
		HTTP: &HTTPConfig{
			AccessLogLevel: hclog.Info.String(),
			Binds: []*BindConfig{
				{
					Addr: "http://127.0.0.1:8080",
				},
			},
		},
		State: state.DefaultConfig(),
	}
}
