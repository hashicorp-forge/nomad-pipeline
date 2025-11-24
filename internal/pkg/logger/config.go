package logger

import (
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"

	"github.com/hashicorp/nomad/helper/pointer"
)

type Config struct {
	Level            string `hcl:"level,optional"`
	JSON             *bool  `hcl:"json,optional"`
	IncludeLine      *bool  `hcl:"include_line,optional"`
	EnableStacktrace *bool  `hcl:"enable_stacktrace,optional"`
}

func DefaultControlerConfig() *Config {
	return &Config{
		Level:            zap.InfoLevel.String(),
		JSON:             pointer.Of(false),
		IncludeLine:      pointer.Of(false),
		EnableStacktrace: pointer.Of(false),
	}
}

func DefaultRunnerConfig() *Config {
	return &Config{
		Level:            zap.InfoLevel.String(),
		JSON:             pointer.Of(true),
		IncludeLine:      pointer.Of(true),
		EnableStacktrace: pointer.Of(true),
	}
}

func (c *Config) Merge(other *Config) *Config {

	if c == nil {
		return other
	}

	result := *c

	if other.Level != "" {
		result.Level = other.Level
	}
	if other.JSON != nil {
		result.JSON = other.JSON
	}
	if other.IncludeLine != nil {
		result.IncludeLine = other.IncludeLine
	}
	if other.EnableStacktrace != nil {
		result.EnableStacktrace = other.EnableStacktrace
	}

	return &result
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "The threshold level for logging",
		},
		&cli.BoolFlag{
			Name:  "log-json",
			Usage: "If the output should be in JSON format",
		},
		&cli.BoolFlag{
			Name:  "log-include-line",
			Usage: "Include file and line information in each log line",
		},
		&cli.BoolFlag{
			Name:  "log-enable-stacktrace",
			Usage: "Enable stacktrace capturing for error level logs",
		},
	}
}

func ConfigFromCLI(cmd *cli.Command) *Config {
	return &Config{
		Level: cmd.String("log-level"),
		JSON: func() *bool {
			if cmd.IsSet("log-json") {
				val := cmd.Bool("log-json")
				return &val
			}
			return nil
		}(),
		IncludeLine: func() *bool {
			if cmd.IsSet("log-include-line") {
				val := cmd.Bool("log-include-line")
				return &val
			}
			return nil
		}(),
		EnableStacktrace: func() *bool {
			if cmd.IsSet("log-enable-stacktrace") {
				val := cmd.Bool("log-enable-stacktrace")
				return &val
			}
			return nil
		}(),
	}
}
