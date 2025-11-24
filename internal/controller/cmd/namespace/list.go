package namespace

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Category:  "namespace",
		Usage:     "List Nomad Pipeline namespaces",
		UsageText: "nomad-pipeline namespace list [options]",
		Flags:     helper.ClientFlags(false),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 0 {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, fmt.Errorf("expected 0 arguments, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.NamespaceListReq{}

			resp, _, err := client.Namespaces().List(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(listCommandCLIErrorMsg, err), 1)
			}

			outputNamespaceList(cmd, resp.Namespaces)

			return nil
		},
	}
}

func outputNamespaceList(cmd *cli.Command, namespaces []*api.NamespaceStub) {
	if len(namespaces) == 0 {
		_, _ = fmt.Fprint(cmd.Writer, "No namespaces found\n")
		return
	}

	out := pterm.TableData{{"ID", "Description"}}

	for _, ns := range namespaces {
		out = append(out, []string{
			ns.ID,
			ns.Description,
		})
	}

	_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
}
