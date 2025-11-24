package trigger

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
		Category:  "trigger",
		Usage:     "List Nomad Pipeline triggers",
		UsageText: "nomad-pipeline trigger list [options]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.TriggerListReq{}

			resp, _, err := client.Triggers().List(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputTriggerList(cmd, resp.Triggers)
			return nil
		},
	}
}

func outputTriggerList(cmd *cli.Command, triggers []*api.TriggerStub) {
	if len(triggers) == 0 {
		_, _ = fmt.Fprint(cmd.Writer, "No triggers found\n")
		return
	}

	out := pterm.TableData{{"ID", "Namespace", "Flow"}}

	for _, trigger := range triggers {
		out = append(out, []string{
			trigger.ID,
			trigger.Namespace,
			trigger.Flow,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
}
