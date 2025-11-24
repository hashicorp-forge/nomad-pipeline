package namespace

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "namespace",
		Usage:     "Get a Nomad Pipeline namespace",
		UsageText: "nomad-pipeline flow get [options] [namespace-id]",
		Flags:     helper.ClientFlags(false),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.NamespaceGetReq{Name: cmd.Args().First()}

			resp, _, err := client.Namespaces().Get(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputNamesapce(resp.Namespace)
			return nil
		},
	}
}

func outputNamesapce(namespace *api.Namespace) {
	pterm.DefaultBasicText.Println(helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", namespace.ID),
		fmt.Sprintf("Description|%s", namespace.Description),
	}))
}
