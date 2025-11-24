package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/version"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner/cmd/job"
)

func main() {

	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprint(cmd.Writer, helper.FormatKV([]string{
			fmt.Sprintf("Version|%s", cmd.Version),
			fmt.Sprintf("Build Time|%s", version.BuildTime),
			fmt.Sprintf("Build Commit|%s", version.BuildCommit),
		}))
		_, _ = fmt.Fprint(cmd.Writer, "\n")
	}

	cliApp := cli.Command{
		Commands: []*cli.Command{
			job.Command(),
		},
		Name:  "nomad-pipeline-runner",
		Usage: "The host runner for Nomad Pipeline",
		Description: strings.TrimSpace(`
Nomad Pipeline Runner executes flow jobs on the host machine as directed
by the Nomad Pipeline controller. It will send status updates and logs
back to the controller via RPC calls to facilitate monitoring and
orchestration`),
		Version:         version.Get(),
		HideHelpCommand: true,
	}

	if err := cliApp.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
