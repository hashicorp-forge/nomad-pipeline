package run

import (
	"github.com/urfave/cli/v2"
)

const (
	getCommandCLIErrorMsg     = "failed to get Nomad Pipeline run detail"
	listCommandCLIErrorMsg    = "failed to list Nomad Pipeline runs"
	logsCommandCLIErrorMsg    = "failed to get Nomad Pipeline run logs"
	monitorCommandCLIErrorMsg = "failed to monitor Nomad Pipeline run"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "run",
		Usage:           "Read and delete Nomad Pipeline runs",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline run <command> [options] [args]",
		Subcommands: []*cli.Command{
			getCommand(),
			listCommand(),
			logsCommand(),
			monitorCommand(),
		},
	}
}
