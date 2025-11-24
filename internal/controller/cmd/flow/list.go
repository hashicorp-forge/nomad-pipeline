package flow

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
		Category:  "flow",
		Usage:     "List Nomad Pipeline flows",
		UsageText: "nomad-pipeline flow list [options]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.FlowListReq{}

			resp, _, err := client.Flows().List(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputFlowList(cmd, resp.Flows)
			return nil
		},
	}
}

func outputFlowList(cmd *cli.Command, flows []*api.FlowStub) {
	if len(flows) == 0 {
		_, _ = fmt.Fprint(cmd.Writer, "No flows found\n")
		return
	}

	out := pterm.TableData{{"ID", "Namespace", "Type"}}

	for _, flow := range flows {
		out = append(out, []string{
			flow.ID,
			flow.Namespace,
			flow.Type,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
}
