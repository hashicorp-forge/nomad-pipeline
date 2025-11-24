package run

import (
	"context"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func logsCommand() *cli.Command {
	return &cli.Command{
		Name:      "logs",
		Category:  "run",
		Usage:     "Get the logs of a Nomad Pipeline run",
		UsageText: "nomad-pipeline run logs [options] [run-id]",
		Flags:     append(helper.ClientFlags(true), logsCommandFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
			}

			if cmd.Bool("tail") {
				return logStream(ctx, cmd, id)
			}
			return logGet(ctx, cmd, id)
		},
	}
}

func logsCommandFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "step-id",
			Required: true,
			Value:    "",
			Usage:    "The flow step ID to get logs for",
		},
		&cli.StringFlag{
			Name:  "type",
			Value: "stdout",
			Usage: "The log type to get (stdout or stderr)",
		},
		&cli.BoolFlag{
			Name:  "tail",
			Value: false,
			Usage: "Whether to tail the logs or not",
		},
	}
}

func logStream(ctx context.Context, cmd *cli.Command, runID ulid.ULID) error {

	req := api.RunLogsTailReq{
		ID:     runID,
		StepID: cmd.String("step-id"),
		Type:   cmd.String("type"),
	}

	client := api.NewClient(helper.ClientConfigFromFlags(cmd))

	resp, _, err := client.Runs().LogsTail(ctx, &req)
	if err != nil {
		return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-resp.ErrCh:
			return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
		case line := <-resp.LogCh:
			_, _ = fmt.Fprint(cmd.Writer, line+"\n")
		}
	}
}

func logGet(ctx context.Context, cmd *cli.Command, runID ulid.ULID) error {

	req := api.RunLogsGetReq{
		ID:     runID,
		JobID:  cmd.String("job-id"),
		StepID: cmd.String("step-id"),
		Type:   cmd.String("type"),
	}

	client := api.NewClient(helper.ClientConfigFromFlags(cmd))

	resp, _, err := client.Runs().LogsGet(ctx, &req)
	if err != nil {
		return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
	}

	for _, log := range resp.Logs {
		_, _ = fmt.Fprint(cmd.Writer, log+"\n")
	}
	return nil
}
