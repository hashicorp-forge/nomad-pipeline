package run

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Category:  "run",
		Usage:     "List Nomad Pipeline runs",
		UsageText: "nomad-pipeline run list [options]",
		Flags:     helper.ClientFlags(true),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg,
					fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.RunListReq{}

			resp, _, err := client.Runs().List(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputRunList(cmd, resp.Runs)
			return nil
		},
	}
}

func outputRunList(cmd *cli.Command, runs []*api.RunStub) {
	if len(runs) == 0 {
		_, _ = fmt.Fprint(cmd.Writer, "No runs found\n")
		return
	}

	out := pterm.TableData{{"ID", "Namespace", "Flow ID", "Status", "CreateTime"}}

	for _, run := range runs {
		out = append(out, []string{
			run.ID.String(),
			run.Namespace,
			run.FlowID,
			colouredRunStatus(run.Status),
			run.CreateTime.String(),
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
}
