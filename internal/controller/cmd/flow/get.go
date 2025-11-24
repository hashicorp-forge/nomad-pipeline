package flow

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Category:  "flow",
		Usage:     "Get a Nomad Pipeline flow",
		UsageText: "nomad-pipeline flow get [options] [flow-id]",
		Flags:     helper.ClientFlags(helper.ClientFlagsWithNamespace),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			req := api.FlowsGetReq{ID: cmd.Args().First()}

			resp, _, err := client.Flows().Get(ctx, &req)
			if err != nil {
				return cli.Exit(helper.FormatError(getCommandCLIErrorMsg, err), 1)
			}

			outputFlow(resp.Flow)
			return nil
		},
	}
}

func outputFlow(f *api.Flow) {

	pterm.DefaultBasicText.Print(helper.FormatKV([]string{
		fmt.Sprintf("ID|%s", f.ID),
		fmt.Sprintf("Namespace|%s", f.Namespace),
		fmt.Sprintf("Type|%s", f.Type()),
	}))
	pterm.DefaultBasicText.Print("\n")

	if len(f.Variables) > 0 {
		out := pterm.TableData{{"Name", "Type", "Default", "Required"}}

		for _, v := range f.Variables {
			out = append(out, []string{
				v.Name,
				variableTypeString(v.Type),
				variableDefaultString(v.Default),
				strconv.FormatBool(v.Required),
			})
		}

		pterm.DefaultSection.Print("Variables")
		_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
	}

	switch f.Type() {
	case api.FlowTypeInline:

		if f.Inline.Runner != nil {
			pterm.DefaultSection.Print(f.Inline.ID + "::" + "Runner")

			pterm.DefaultBasicText.Print(helper.FormatKV([]string{
				fmt.Sprintf("Type|%s", "Nomad on Demand"),
				fmt.Sprintf("Namespace|%s", f.Inline.Runner.NomadOnDemand.Namespace),
				fmt.Sprintf("Image|%s", f.Inline.Runner.NomadOnDemand.Image),
				fmt.Sprintf("CPU MHz|%v", f.Inline.Runner.NomadOnDemand.Resource.CPU),
				fmt.Sprintf("Memory MB|%v", f.Inline.Runner.NomadOnDemand.Resource.Memory),
			}))
			pterm.DefaultBasicText.Print("\n")

			if len(f.Inline.Runner.NomadOnDemand.Artifacts) > 0 {
				pterm.DefaultBasicText.Print("\n")

				out := pterm.TableData{{"Source", "Dest", "Options"}}

				for _, artifact := range f.Inline.Runner.NomadOnDemand.Artifacts {
					out = append(out, []string{
						artifact.Source,
						artifact.Dest,
						artifactOptionsString(artifact.Options),
					})
				}
				_ = pterm.DefaultTable.WithHasHeader().WithData(out).Render()
			}
		}

		for _, step := range f.Inline.Steps {
			pterm.DefaultSection.Print(f.Inline.ID, "::", step.ID)
			if step.Condition != "" {
				pterm.DefaultBasicText.Print(helper.FormatKV([]string{
					fmt.Sprintf("Conditional|%s", step.Condition),
				}))
				pterm.DefaultBasicText.Print("\n")
			}
			pterm.DefaultBox.Println(step.Run)
		}
	case api.FlowTypeSpecification:
		for _, spec := range f.Specification {
			pterm.DefaultSection.Print(spec.ID)
			if spec.Condition != "" {
				pterm.Println(fmt.Sprintf("Condition: %q", spec.Condition))
			}
			if spec.Job.NameFormat != "" {
				pterm.Println(fmt.Sprintf("Job Name Format: %q", spec.Job.NameFormat))
			}
			pterm.DefaultBox.Println(spec.Job.Raw)
		}
	}
}

func variableTypeString(t string) string {
	if t == "" {
		return "<any>"
	}
	return t
}

func variableDefaultString(v any) string {
	if v == nil {
		return "<none>"
	}
	return fmt.Sprintf("%v", v)
}

func artifactOptionsString(opts map[string]string) string {

	if len(opts) == 0 {
		return "<none>"
	}

	kvs := make([]string, 0, len(opts))

	for key, value := range opts {
		kvs = append(kvs, key+"="+value)
	}
	return strings.Join(kvs, ",")
}
