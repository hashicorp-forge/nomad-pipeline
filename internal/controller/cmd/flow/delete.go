package flow

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func deleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Category:  "flow",
		Usage:     "Delete a Nomad Pipeline flow",
		UsageText: "nomad-pipeline flow delete [options] [flow-id]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(deleteCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.FlowDeleteReq{ID: cmd.Args().First()}

			_, err := client.Flows().Delete(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(deleteCommandCLIErrorMsg, err), 1)
			}

			_, _ = fmt.Fprintf(cmd.Writer, "successfully deleted Nomad Pipeline flow %q\n", cmd.Args().First())
			return nil
		},
	}
}
