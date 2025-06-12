package logger

import "github.com/hashicorp/go-hclog"

func New(cfg *Config) hclog.Logger {
	return hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:            cfg.Name,
		Level:           hclog.LevelFromString(cfg.Level),
		JSONFormat:      *cfg.JSON,
		IncludeLocation: *cfg.IncludeLine,
	})
}
