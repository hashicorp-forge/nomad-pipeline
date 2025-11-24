package flow

import (
	"github.com/urfave/cli/v3"
)

const (
	createCommandCLIErrorMsg = "failed to create Nomad Pipeline flow"
	deleteCommandCLIErrorMsg = "failed to delete Nomad Pipeline flow"
	getCommandCLIErrorMsg    = "failed to get Nomad Pipeline flow"
	listCommandCLIErrorMsg   = "failed to list Nomad Pipeline flows"
	runCommandCLIErrorMsg    = "failed to run Nomad Pipeline flow"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "flow",
		Usage:           "Create, read, delete, and run Nomad Pipeline flows",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline flow <command> [options] [args]",
		Commands: []*cli.Command{
			createCommand(),
			deleteCommand(),
			getCommand(),
			listCommand(),
			runCommand(),
		},
	}
}
