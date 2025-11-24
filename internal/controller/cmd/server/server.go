package server

import (
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "server",
		Usage:           "Run, control, and interrogate Nomad Pipeline servers",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline server <command> [options] [args]",
		Commands: []*cli.Command{
			runCommand(),
		},
	}
}
