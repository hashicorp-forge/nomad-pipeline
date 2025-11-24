package flow

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Category:  "flow",
		Usage:     "Create a Nomad Pipeline flow",
		UsageText: "nomad-pipeline flow create [options] [flow-spec]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			flow, err := api.ParseFlowFile(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.FlowCreateReq{Flow: flow}

			resp, _, err := client.Flows().Create(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			outputFlow(resp.Flow)
			return nil
		},
	}
}
