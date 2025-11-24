package trigger

import (
	"context"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "trigger",
		Usage:     "Get a Nomad Pipeline trigger",
		UsageText: "nomad-pipeline trigger get [options] [trigger-id]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.TriggersGetReq{ID: cmd.Args().First()}

			resp, _, err := client.Triggers().Get(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputTrigger(resp.Trigger)
			return nil
		},
	}
}

func outputTrigger(t *api.Trigger) {
	pterm.DefaultBasicText.Print(helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", t.ID),
		fmt.Sprintf("Namespace|%s", t.Namespace),
		fmt.Sprintf("Flow|%s", t.Flow),
		fmt.Sprintf("Source ID|%s", t.Source.ID),
		fmt.Sprintf("Source Provider|%s", t.Source.Provider),
	}))
	pterm.DefaultBasicText.Print("\n")

	pterm.DefaultSection.Print("Source Configuration")
	pterm.DefaultBasicText.Println(strings.ReplaceAll(string(t.Source.Config), " ", ""))
}
