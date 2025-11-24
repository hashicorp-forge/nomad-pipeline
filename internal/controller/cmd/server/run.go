package server

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
)

func runCommand() *cli.Command {
	return &cli.Command{
		Name:     "run",
		Category: "server",
		Usage:    "Run a Nomad Pipeline server",
		Flags:    append(logger.Flags(), server.Flags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			defaultCfg := server.DefaultConfig()

			// Merge logger config from CLI
			defaultCfg.Log = defaultCfg.Log.Merge(logger.ConfigFromCLI(cmd))

			// Merge server config from CLI
			defaultCfg = defaultCfg.Merge(server.ConfigFromCLI(cmd))

			srv, err := server.NewServer(defaultCfg)
			if err != nil {
				return cli.Exit(helper.FormatError("failed to run an Nomad Pipeline server", err), 1)
			}
			srv.Start()
			srv.WaitForSignals()
			return nil
		},
	}
}
