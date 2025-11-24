package run

import (
	"github.com/urfave/cli/v3"
)

const (
	cancelCommandCLIErrorMsg  = "failed to cancel Nomad Pipeline run detail"
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
		Commands: []*cli.Command{
			cancelCommand(),
			getCommand(),
			listCommand(),
			logsCommand(),
			monitorCommand(),
		},
	}
}
