package namespace

import (
	"context"
	"fmt"

	cliHelper "github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
	"github.com/urfave/cli/v3"
)

func deleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Category:  "namespace",
		Usage:     "Delete a Nomad Pipeline namespace",
		UsageText: "nomad-pipeline namespace delete [options] [namespace-id]",
		Flags:     cliHelper.ClientFlags(false),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(cliHelper.FormatError(deleteCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(cliHelper.ClientConfigFromFlags(cmd))

			req := api.NamespaceDeleteReq{Name: cmd.Args().First()}

			_, err := client.Namespaces().Delete(ctx, &req)
			if err != nil {
				return cli.Exit(cliHelper.FormatError(deleteCommandCLIErrorMsg, err), 1)
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Namespace '%s' deleted successfully\n", req.Name)

			return nil
		},
	}
}
