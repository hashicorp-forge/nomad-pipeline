package trigger

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
		Category:  "trigger",
		Usage:     "Delete a Nomad Pipeline trigger",
		UsageText: "nomad-pipeline trigger delete [options] [trigger-id]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(deleteCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.TriggerDeleteReq{ID: cmd.Args().First()}

			_, err := client.Triggers().Delete(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(deleteCommandCLIErrorMsg, err), 1)
			}

			_, _ = fmt.Fprintf(cmd.Writer, "successfully deleted Nomad Pipeline trigger %q\n", cmd.Args().First())
			return nil
		},
	}
}
