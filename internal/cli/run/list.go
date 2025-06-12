package run

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Category:  "run",
		Usage:     "List Nomad Pipeline runs",
		Args:      false,
		UsageText: "nomad-pipeline run list [options]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg,
					fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.RunListReq{}

			resp, _, err := client.Runs().List(context.Background(), &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputRunList(cliCtx, resp.Runs)
			return nil
		},
	}
}

func outputRunList(cliCtx *cli.Context, runs []*api.RunStub) {
	if len(runs) == 0 {
		_, _ = fmt.Fprint(cliCtx.App.Writer, "No runs found\n")
		return
	}

	out := make([]string, 0, len(runs)+1)
	out = append(out, "ID|Flow ID|Status|Start Time|End Time")
	for _, run := range runs {
		out = append(out, fmt.Sprintf(
			"%s|%s|%s|%v|%v",
			run.ID, run.FlowID, run.Status, helper.FormatTime(run.StartTime), helper.FormatTime(run.EndTime)))
	}

	_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatList(out))
	_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n")
}
