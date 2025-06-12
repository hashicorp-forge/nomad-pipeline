package flow

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Category:  "flow",
		Usage:     "Create a Nomad Pipeline flow",
		Args:      true,
		UsageText: "nomad-pipeline flow create [options] [flow-spec]",
		Flags:     helper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			flow, err := api.ParseFlowFile(cliCtx.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cliCtx))

			req := api.FlowCreateReq{Flow: flow}

			resp, _, err := client.Flows().Create(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			outputFlow(cliCtx, resp.Flow)
			return nil
		},
	}
}
