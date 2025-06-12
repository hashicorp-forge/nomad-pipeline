package flow

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Category:  "flow",
		Usage:     "List Nomad Pipeline flows",
		Args:      false,
		UsageText: "nomad-pipeline flow list [options]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.FlowListReq{}

			resp, _, err := client.Flows().List(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputFlowList(cliCtx, resp.Flows)
			return nil
		},
	}
}

func outputFlowList(cliCtx *cli.Context, flows []*api.FlowStub) {
	if len(flows) == 0 {
		_, _ = fmt.Fprint(cliCtx.App.Writer, "No flows found\n")
		return
	}

	out := make([]string, 0, len(flows)+1)
	out = append(out, "ID|Jobs")
	for _, flow := range flows {
		out = append(out, fmt.Sprintf(
			"%s|%s",
			flow.ID, strings.Join(flow.Jobs, ", ")))
	}

	_, _ = fmt.Fprint(cliCtx.App.Writer, helper.FormatList(out))
	_, _ = fmt.Fprintf(cliCtx.App.Writer, "\n")
}
