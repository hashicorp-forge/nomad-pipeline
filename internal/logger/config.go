package logger

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/helper/pointer"
)

type Config struct {
	Name        string `hcl:"name,optional"`
	Level       string `hcl:"level,optional"`
	JSON        *bool  `hcl:"json,optional"`
	IncludeLine *bool  `hcl:"include_line,optional"`
}

func DefaultConfig() *Config {
	return &Config{
		Name:        Name,
		Level:       hclog.Info.String(),
		JSON:        pointer.Of(false),
		IncludeLine: pointer.Of(false),
	}
}

const (
	Name = "nomad-pipeline"

	ComponentNameServer = "server"
	ComponentNameHTTP   = "http"
	ComponentNameRunner = "runner"
)
