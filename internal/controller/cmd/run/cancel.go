package run

import (
	"context"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func cancelCommand() *cli.Command {
	return &cli.Command{
		Name:      "cancel",
		Category:  "run",
		Usage:     "Cancel Nomad Pipeline runs",
		UsageText: "nomad-pipeline run cancel [options] [run-id]",
		Flags:     helper.ClientFlags(true),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(cancelCommandCLIErrorMsg,
					fmt.Errorf("expected 1 arguments, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(cancelCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.RunCancelReq{ID: id}

			_, _, err = client.Runs().Cancel(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(cancelCommandCLIErrorMsg, err), 1)
			}

			pterm.DefaultBasicText.Printf("Run '%s' cancelled successfully", id)
			return nil
		},
	}
}
