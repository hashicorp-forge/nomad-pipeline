package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/flow"
	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/run"
	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/server"
	"github.com/hashicorp-forge/nomad-pipeline/internal/version"
)

func main() {

	cli.VersionPrinter = func(cliCtx *cli.Context) {
		_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatKV([]string{
			fmt.Sprintf("Version|%s", cliCtx.App.Version),
			fmt.Sprintf("Build Time|%s", version.BuildTime),
			fmt.Sprintf("Build Commit|%s", version.BuildCommit),
		}))
		_, _ = fmt.Fprint(cliCtx.App.Writer, "\n")
	}

	cliApp := cli.App{
		Commands: []*cli.Command{
			flow.Command(),
			run.Command(),
			server.Command(),
		},
		Name:  "nomad-pipeline",
		Usage: "Job chaining and pipeline orchestration for Nomad",
		Description: `Nomad Pipeline augments Nomad by allowing declarative job chaining and
pipeline orchestration. Whether you're running ML workloads, ETL jobs, or
multi-stage compute tasks, Nomad Pipeline streamlines pipeline execution
and monitoring within your existing Nomad infrastructure.`,
		Version:         version.Get(),
		HideHelpCommand: true,
	}

	if err := cliApp.Run(os.Args); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
