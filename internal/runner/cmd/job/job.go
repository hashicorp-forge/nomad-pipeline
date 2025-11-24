package job

import "github.com/urfave/cli/v3"

func Command() *cli.Command {
	return &cli.Command{
		Name:            "job",
		Usage:           "Execute a Nomad Pipeline flow job",
		HideHelpCommand: true,
		UsageText:       "nomad-pipeline-runner job <command> [options] [args]",
		Commands: []*cli.Command{
			runCommand(),
		},
	}
}
