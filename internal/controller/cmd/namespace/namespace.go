package namespace

import "github.com/urfave/cli/v3"

const (
	createCommandCLIErrorMsg = "failed to create Nomad Pipeline namespace"
	deleteCommandCLIErrorMsg = "failed to delete Nomad Pipeline namespace"
	getCommandCLIErrorMsg    = "failed to get Nomad Pipeline namespace"
	listCommandCLIErrorMsg   = "failed to list Nomad Pipeline namespace"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "namespace",
		Usage:           "Create, read, and delete Nomad Pipeline namespaces",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline namespace <command> [options] [args]",
		Commands: []*cli.Command{
			createCommand(),
			deleteCommand(),
			getCommand(),
			listCommand(),
		},
	}
}
