package namespace

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	cliHelper "github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/hcl"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Category:  "namespace",
		Usage:     "Create a Nomad Pipeline namespace",
		UsageText: "nomad-pipeline namespace create [options] [namespace-spec]",
		Flags:     cliHelper.ClientFlags(false),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(cliHelper.FormatError(createCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			decodeObj := struct {
				Namespace *api.Namespace `hcl:"namespace,block"`
			}{}

			if err := hcl.ParseConfig(cmd.Args().First(), &decodeObj); err != nil {
				return cli.Exit(cliHelper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(cliHelper.ClientConfigFromFlags(cmd))

			req := api.NamespaceCreateReq{Namespace: decodeObj.Namespace}

			resp, _, err := client.Namespaces().Create(ctx, &req)
			if err != nil {
				return cli.Exit(cliHelper.FormatError(createCommandCLIErrorMsg, err), 1)
			}

			outputNamesapce(resp.Namespace)
			return nil
		},
	}
}
