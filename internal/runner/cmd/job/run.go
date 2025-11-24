package job

import (
	"context"

	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/job"
	"github.com/urfave/cli/v3"
)

func runCommandFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Required: true,
			Value:    "",
			Usage:    "Path to the run configuration",
		},
	}
}

func runCommand() *cli.Command {
	return &cli.Command{
		Name:     "run",
		Category: "job",
		Usage:    "Run a Nomad Pipeline flow job via the host runner",
		Flags:    runCommandFlags(),
		Action: func(_ context.Context, cmd *cli.Command) error {

			runner, err := job.NewRunner(cmd.String("config"))
			if err != nil {
				return cli.Exit(err, 1)
			}

			if err := runner.Run(); err != nil {
				return cli.Exit(err, 1)
			}
			return nil
		},
	}
}
