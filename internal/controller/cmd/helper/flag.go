package helper

import (
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

const (
	addressCLIFlag = "address"
	namespaceFlag  = "namespace"

	ClientFlagsWithNamespace    = true
	ClientFlagsWithoutNamespace = false
)

func ClientFlags(namespace bool) []cli.Flag {
	f := []cli.Flag{
		&cli.StringFlag{
			Aliases: []string{"a"},
			Sources: cli.EnvVars("NOMAD_PIPELINE_ADDR"),
			Name:    addressCLIFlag,
			Value:   "http://127.0.0.1:8080",
			Usage:   "Nomad Pipeline server address to make API requests to",
		},
	}

	if namespace {
		f = append(f, &cli.StringFlag{
			Aliases: []string{"n"},
			Sources: cli.EnvVars("NOMAD_PIPELINE_NAMESPACE"),
			Name:    namespaceFlag,
			Value:   "default",
			Usage:   "Nomad Pipeline namespace to make API requests to",
		})
	}

	return f
}

func ClientConfigFromFlags(cmd *cli.Command) *api.Config {

	defaultConfig := api.DefaultConfig()

	if addr := cmd.String(addressCLIFlag); addr != "" {
		defaultConfig.Address = addr
	}
	if namespace := cmd.String(namespaceFlag); namespace != "" {
		defaultConfig.Namespace = namespace
	}

	return defaultConfig
}
