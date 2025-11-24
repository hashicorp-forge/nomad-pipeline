package run

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/cmd/helper"
	"github.com/hashicorp-forge/nomad-pipeline/pkg/api/v1"
)

func monitorCommand() *cli.Command {
	return &cli.Command{
		Name:      "monitor",
		Category:  "run",
		Usage:     "Monitor a Nomad Pipeline run",
		UsageText: "nomad-pipeline run monitor [options] [run-id]",
		Flags:     helper.ClientFlags(true),
		Action: func(ctx context.Context, cmd *cli.Command) error {

			if numArgs := cmd.Args().Len(); numArgs != 1 {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg,
					fmt.Errorf("expected 1 argument, got %v", numArgs)), 1)
			}

			id, err := ulid.Parse(cmd.Args().First())
			if err != nil {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg, err), 1)
			}

			client := api.NewClient(helper.ClientConfigFromFlags(cmd))

			resp, _, err := client.Runs().Get(ctx, &api.RunGetReq{ID: id})
			if err != nil {
				return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg, err), 1)
			}

			switch resp.Run.Status {
			case api.RunStatusFailed, api.RunStatusSuccess, api.RunStatusCancelled:
				outputRun(resp.Run)
			default:
				monitorRun(ctx, client, resp.Run)
			}
			return nil
		},
	}
}

func MonitorRun(ctx context.Context, client *api.Client, runID ulid.ULID) error {
	resp, _, err := client.Runs().Get(ctx, &api.RunGetReq{ID: runID})
	if err != nil {
		return cli.Exit(helper.FormatError(monitorCommandCLIErrorMsg, err), 1)
	}
	monitorRun(ctx, client, resp.Run)
	return nil
}

func monitorRun(ctx context.Context, client *api.Client, run *api.Run) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	consoleUpdater := &runUpdater{
		area:      &pterm.AreaPrinter{},
		logBuffer: &logBuffer{lines: make([]string, 0, 10)},
	}
	consoleUpdater.start()
	defer consoleUpdater.stop()

	// Start log streaming if we have inline steps
	if run.InlineRun != nil && len(run.InlineRun.Steps) > 0 {
		for _, step := range run.InlineRun.Steps {
			if step.Status == api.RunStatusRunning || step.Status == api.RunStatusPending {
				go consoleUpdater.streamLogs(ctx, client, run.ID, step.ID)
				break
			}
		}
	}

	consoleUpdater.update(run)

	for {
		select {
		case <-ticker.C:
			resp, _, err := client.Runs().Get(ctx, &api.RunGetReq{ID: run.ID})
			if err != nil {
				continue
			}

			consoleUpdater.update(resp.Run)

			// Check if we need to start streaming logs from a new step
			if resp.Run.InlineRun != nil {
				for _, step := range resp.Run.InlineRun.Steps {
					if step.Status == api.RunStatusRunning && !consoleUpdater.isStreamingStep(step.ID) {
						go consoleUpdater.streamLogs(ctx, client, resp.Run.ID, step.ID)
						break
					}
				}
			}

			if resp.Run.Status == api.RunStatusFailed || resp.Run.Status == api.RunStatusSuccess || resp.Run.Status == api.RunStatusCancelled {
				consoleUpdater.update(resp.Run)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

type logBuffer struct {
	mu    sync.RWMutex
	lines []string
}

func (lb *logBuffer) add(line string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.lines = append(lb.lines, line)
	if len(lb.lines) > 10 {
		lb.lines = lb.lines[len(lb.lines)-10:]
	}
}

func (lb *logBuffer) get() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	result := make([]string, len(lb.lines))
	copy(result, lb.lines)
	return result
}

type runUpdater struct {
	area            *pterm.AreaPrinter
	logBuffer       *logBuffer
	currentStepID   string
	currentStepMu   sync.RWMutex
	cancelLogStream context.CancelFunc
}

func (ru *runUpdater) start() {
	_ = ru.area.Stop()
}

func (ru *runUpdater) stop() {
	_ = ru.area.Stop()
	if ru.cancelLogStream != nil {
		ru.cancelLogStream()
	}
}

func (ru *runUpdater) isStreamingStep(stepID string) bool {
	ru.currentStepMu.RLock()
	defer ru.currentStepMu.RUnlock()
	return ru.currentStepID == stepID
}

func (ru *runUpdater) streamLogs(ctx context.Context, client *api.Client, runID ulid.ULID, stepID string) {
	ru.currentStepMu.Lock()
	if ru.cancelLogStream != nil {
		ru.cancelLogStream()
	}

	streamCtx, cancel := context.WithCancel(ctx)
	ru.cancelLogStream = cancel
	ru.currentStepID = stepID
	ru.currentStepMu.Unlock()

	req := api.RunLogsTailReq{
		ID:     runID,
		StepID: stepID,
		Type:   "stdout",
	}

	resp, _, err := client.Runs().LogsTail(streamCtx, &req)
	if err != nil {
		return
	}

	for {
		select {
		case <-streamCtx.Done():
			return
		case <-resp.ErrCh:
			return
		case line := <-resp.LogCh:
			ru.logBuffer.add(line)
		}
	}
}

func (ru *runUpdater) update(run *api.Run) {
	content := ru.buildContent(run)
	ru.area.Update(content)
}

func (ru *runUpdater) buildContent(run *api.Run) string {
	var out string
	out += pterm.DefaultBasicText.Sprint(runHeader(run))
	out += "\n"

	if runVars := runVariables(run); runVars != "" {
		out += pterm.DefaultSection.Sprint("Variables")
		out += pterm.DefaultBasicText.Sprint(runVars)
	}

	out += pterm.DefaultSection.Sprint("Steps")
	out += pterm.DefaultBasicText.Sprint(runBody(run))

	// Add log section at the bottom
	logLines := ru.logBuffer.get()
	if len(logLines) > 0 {
		out += pterm.DefaultSection.Sprint("Log Stream")
		out += ru.formatLogs(logLines)
	}

	return out
}

func (ru *runUpdater) formatLogs(lines []string) string {

	// Get terminal width for truncation, so log lines don't wrap around and
	// mess up the display area.
	terminalWidth := pterm.GetTerminalWidth()
	if terminalWidth <= 0 {
		terminalWidth = pterm.FallbackTerminalWidth
	}

	var out strings.Builder
	for _, line := range lines {
		if len(line) > terminalWidth {
			out.WriteString(line[:terminalWidth])
		} else {
			out.WriteString(line)
		}
		out.WriteString("\n")
	}
	return out.String()
}
