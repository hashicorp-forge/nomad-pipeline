package trigger

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
		Category:  "trigger",
		Usage:     "Create a Nomad Pipeline trigger",
		UsageText: "nomad-pipeline trigger create [options] [trigger-spec]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			triggerSpec, err := api.ParseTriggerFile(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			if err := triggerSpec.Validate(); err != nil {
				return cli.Exit(
					helper.FormatError(
						createCommandCLIErrorMsg,
						fmt.Errorf("invalid trigger spec: %v", err),
					),
					1,
				)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.TriggerCreateReq{Trigger: triggerSpec}

			resp, _, err := client.Triggers().Create(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			outputTrigger(resp.Trigger)
			return nil
		},
	}
}
