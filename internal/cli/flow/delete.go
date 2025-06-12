package flow

import (
	"fmt"

	"github.com/urfave/cli/v2"

	cliHelper "github.com/hashicorp-forge/nomad-pipeline/internal/cli/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api"
)

func deleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Category:  "flow",
		Usage:     "Delete a Nomad Pipeline flow",
		Args:      true,
		UsageText: "nomad-pipeline flow delete [options] [flow-id]",
		Flags:     cliHelper.ClientFlags(),
		Action: func(cliCtx *cli.Context) error {

			if numArgs := cliCtx.Args().Len(); numArgs != 1 {
				return cli.Exit(cliHelper.FormatError(deleteCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(cliHelper.ClientConfigFromFlags(cliCtx))

			req := api.FlowDeleteReq{ID: cliCtx.Args().First()}

			_, err := client.Flows().Delete(cliCtx.Context, &req)
			if err != nil {
				return cli.Exit(cliHelper.FormatError(deleteCommandCLIErrorMsg, err), 1)
			}

			_, _ = fmt.Fprintf(cliCtx.App.Writer, "successfully deleted Nomad Pipeline flow %q\n", cliCtx.Args().First())
			return nil
		},
	}
}
