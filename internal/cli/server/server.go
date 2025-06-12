package server

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "server",
		Usage:           "Run, control, and interrogate Nomad Pipeline servers",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline server <command> [options] [args]",
		Subcommands: []*cli.Command{
			runCommand(),
		},
	}
}
