package helper

import (
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

const (
	addressCLIFlag = "address"
)

func ClientFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Aliases: []string{"a"},
			Name:    addressCLIFlag,
			Value:   "http://127.0.0.1:8080",
			Usage:   "Nomad Pipeline server address to make API requests to",
		},
	}
}

func ClientConfigFromFlags(ctx *cli.Context) *api.Config {

	defaultConfig := api.DefaultConfig()

	if addr := ctx.String(addressCLIFlag); addr != "" {
		defaultConfig.Address = addr
	}

	return defaultConfig
}
