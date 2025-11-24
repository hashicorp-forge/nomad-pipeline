package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/flow"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/namespace"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/run"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/server"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/trigger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/version"
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
			flow.Command(),
			namespace.Command(),
			run.Command(),
			server.Command(),
			trigger.Command(),
		},
		Name:  "nomad-pipeline",
		Usage: "Job chaining and pipeline orchestration for Nomad",
		Description: strings.TrimSpace(
`Nomad Pipeline augments Nomad by allowing declarative job chaining and
pipeline orchestration. Whether you're running ML workloads, ETL jobs, or
multi-stage compute tasks, Nomad Pipeline streamlines pipeline execution
and monitoring within your existing Nomad infrastructure.`),
		Version:         version.Get(),
		HideHelpCommand: true,
	}

	if err := cliApp.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
}
