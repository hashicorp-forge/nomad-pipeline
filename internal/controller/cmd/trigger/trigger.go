package trigger

import (
	"github.com/urfave/cli/v3"
)

const (
	createCommandCLIErrorMsg = "failed to create Nomad Pipeline trigger"
	deleteCommandCLIErrorMsg = "failed to delete Nomad Pipeline trigger"
	getCommandCLIErrorMsg    = "failed to get Nomad Pipeline trigger"
	listCommandCLIErrorMsg   = "failed to list Nomad Pipeline triggers"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "trigger",
		Usage:           "Create, read, and delete Nomad Pipeline triggers",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline trigger <command> [options] [args]",
		Commands: []*cli.Command{
			createCommand(),
			deleteCommand(),
			getCommand(),
			listCommand(),
		},
	}
}
