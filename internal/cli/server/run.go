package server

import (
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server"
)

func runCommand() *cli.Command {
	return &cli.Command{
		Name:     "run",
		Category: "server",
		Usage:    "Run a Nomad Pipeline server",
		Action: func(cliCtx *cli.Context) error {

			srv, err := server.NewServer(server.DefaultConfig())
			if err != nil {
				return cli.Exit(helper.FormatError("failed to run an Nomad Pipeline server", err), 1)
			}
			srv.Start()
			srv.WaitForSignals()
			return nil
		},
	}
}
