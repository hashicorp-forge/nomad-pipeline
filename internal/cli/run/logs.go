package run

import (
	"context"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func logsCommandFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "job-id",
			Required: true,
			Value:    "",
			Usage:    "The flow job ID to get logs for",
		},
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

func logsCommand() *cli.Command {
	return &cli.Command{
		Name:      "logs",
		Category:  "run",
		Usage:     "Get the logs of a Nomad Pipeline run",
		Args:      false,
		UsageText: "nomad-pipeline run logs [options] [run-id]",
		Flags:     append(helper.ClientFlags(), logsCommandFlags()...),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cliCtx.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
			}

			if cliCtx.Bool("tail") {
				return logStream(cliCtx, id)
			}
			return logGet(cliCtx, id)
		},
	}
}

func logStream(cliCtx *cli.Context, runID ulid.ULID) error {

	req := api.RunLogsTailReq{
		ID:     runID,
		JobID:  cliCtx.String("job-id"),
		StepID: cliCtx.String("step-id"),
		Type:   cliCtx.String("type"),
	}

	client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

	resp, _, err := client.Runs().LogsTail(context.Background(), &req)
	if err != nil {
		return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
	}

	for {
		select {
		case <-cliCtx.Done():
			return nil
		case err := <-resp.ErrCh:
			return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
		case line := <-resp.LogCh:
			_, _ = fmt.Fprint(cliCtx.App.Writer, line+"\n")
		}
	}
}

func logGet(cliCtx *cli.Context, runID ulid.ULID) error {

	req := api.RunLogsGetReq{
		ID:     runID,
		JobID:  cliCtx.String("job-id"),
		StepID: cliCtx.String("step-id"),
		Type:   cliCtx.String("type"),
	}

	client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

	resp, _, err := client.Runs().LogsGet(cliCtx.Context, &req)
	if err != nil {
		return cli.Exit(helper.FormatError(logsCommandCLIErrorMsg, err), 1)
	}

	for _, log := range resp.Logs {
		_, _ = fmt.Fprint(cliCtx.App.Writer, log+"\n")
	}
	return nil
}
