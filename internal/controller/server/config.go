package server

import (
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
)

type Config struct {
	Data  *DataConfig    `hcl:"data,optional"`
	Log   *logger.Config `hcl:"log,optional"`
	HTTP  *HTTPConfig    `hcl:"http,optional"`
	Nomad *NomadConfig   `hcl:"nomad,optional"`
	RPC   *RPCConfig     `hcl:"rpc,optional"`
	State *state.Config  `hcl:"state,optional"`
}

type DataConfig struct {
	Path string `hcl:"path,optional"`
}

type HTTPConfig struct {
	Addr           string `hcl:"addr,optional"`
	AccessLogLevel string `hcl:"access_log_level,optional"`
}

type RPCConfig struct {
	Addr string `hcl:"addr,optional"`
}

type NomadConfig struct {
	Addr  string `hcl:"addr,optional"`
	Token string `hcl:"token,optional"`
}

func DefaultConfig() *Config {
	return &Config{
		Data: &DataConfig{
			Path: "/tmp/nomad-pipeline/data",
		},
		Log: logger.DefaultControlerConfig(),
		HTTP: &HTTPConfig{
			Addr:           "http://localhost:8080",
			AccessLogLevel: zap.DebugLevel.String(),
		},
		Nomad: &NomadConfig{
			Addr: "http://localhost:4646",
		},
		RPC: &RPCConfig{
			Addr: "http://localhost:8081",
		},
		State: state.DefaultConfig(),
	}
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "data-dir",
			Usage:   "The path to the data directory",
			Sources: cli.EnvVars("NOMAD_PIPELINE_DATA_DIR"),
		},
		&cli.StringFlag{
			Name:    "http-addr",
			Usage:   "The HTTP server address",
			Sources: cli.EnvVars("NOMAD_PIPELINE_HTTP_ADDR"),
		},
		&cli.StringFlag{
			Name:    "http-access-log-level",
			Usage:   "The HTTP access log level (debug, info, warn, error)",
			Sources: cli.EnvVars("NOMAD_PIPELINE_HTTP_ACCESS_LOG_LEVEL"),
		},
		&cli.StringFlag{
			Name:    "nomad-addr",
			Usage:   "The Nomad server address",
			Sources: cli.EnvVars("NOMAD_ADDR"),
		},
		&cli.StringFlag{
			Name:    "nomad-token",
			Usage:   "The Nomad ACL token to use for HTTP requests",
			Sources: cli.EnvVars("NOMAD_TOKEN"),
		},
		&cli.StringFlag{
			Name:    "rpc-addr",
			Usage:   "The RPC server address",
			Sources: cli.EnvVars("NOMAD_PIPELINE_RPC_ADDR"),
		},
		&cli.StringFlag{
			Name:    "state-backend",
			Usage:   "The state backend to use (dev, nomad-vars)",
			Sources: cli.EnvVars("NOMAD_PIPELINE_STATE_BACKEND"),
		},
		&cli.BoolFlag{
			Name:    "state-nomad-vars-cache-enabled",
			Usage:   "Enable Nomad variables state backend read cache",
			Sources: cli.EnvVars("NOMAD_PIPELINE_STATE_NOMAD_VARS_CACHE_ENABLED"),
		},
	}
}

func ConfigFromCLI(cmd *cli.Command) *Config {
	cfg := &Config{
		Data: &DataConfig{
			Path: cmd.String("data-dir"),
		},
		HTTP: &HTTPConfig{
			Addr:           cmd.String("http-addr"),
			AccessLogLevel: cmd.String("http-access-log-level"),
		},
		Nomad: &NomadConfig{
			Addr:  cmd.String("nomad-addr"),
			Token: cmd.String("nomad-token"),
		},
		RPC: &RPCConfig{
			Addr: cmd.String("rpc-addr"),
		},
		State: &state.Config{
			Backend: cmd.String("state-backend"),
			NVars: &state.NomadVarsConfig{
				CacheEnabled: cmd.Bool("state-nomad-vars-cache-enabled"),
			},
		},
	}

	return cfg
}

func (c *Config) Merge(other *Config) *Config {
	if c == nil {
		return other
	}

	result := *c

	if other.Data != nil {
		if result.Data == nil {
			result.Data = &DataConfig{}
		}
		if other.Data.Path != "" {
			result.Data.Path = other.Data.Path
		}
	}

	if other.HTTP != nil {
		if result.HTTP == nil {
			result.HTTP = &HTTPConfig{}
		}
		if other.HTTP.Addr != "" {
			result.HTTP.Addr = other.HTTP.Addr
		}
		if other.HTTP.AccessLogLevel != "" {
			result.HTTP.AccessLogLevel = other.HTTP.AccessLogLevel
		}
	}

	if other.Nomad != nil {
		if result.Nomad == nil {
			result.Nomad = &NomadConfig{}
		}
		if other.Nomad.Addr != "" {
			result.Nomad.Addr = other.Nomad.Addr
		}
		if other.Nomad.Token != "" {
			result.Nomad.Token = other.Nomad.Token
		}
	}

	if other.RPC != nil {
		if result.RPC == nil {
			result.RPC = &RPCConfig{}
		}
		if other.RPC.Addr != "" {
			result.RPC.Addr = other.RPC.Addr
		}
	}

	if other.State != nil {
		if result.State == nil {
			result.State = &state.Config{}
		}
		result.State = result.State.Merge(other.State)
	}

	if other.Log != nil {
		result.Log = result.Log.Merge(other.Log)
	}

	return &result
}
